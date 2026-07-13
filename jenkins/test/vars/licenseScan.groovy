#!/usr/bin/env groovy

def installTools() {
    sh '''
        echo "Installing Trivy..."
        if ! command -v trivy &> /dev/null; then
            curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
        fi
        trivy --version

        echo "Installing Parlay..."
        if ! command -v parlay &> /dev/null; then
            curl -sSfL https://github.com/snyk/parlay/releases/download/v0.9.0/parlay_Linux_x86_64.tar.gz | tar -xz -C /usr/local/bin parlay
            chmod +x /usr/local/bin/parlay
        fi
        parlay --version

        echo "Checking Python3..."
        if ! command -v python3 &> /dev/null; then
            dnf install -y python3
        fi
        python3 --version
    '''
}

def generateTrivySBOM(String imageSource, String trivyFile) {
    if (imageSource.endsWith('.tar')) {
        sh """
            echo "Generating Trivy SBOM from tar file..."
            ls -lh ${imageSource}
            trivy image --input ${imageSource} --format spdx-json --output ${trivyFile}
        """
    } else {
        // Trivy needs ICR credentials to pull from the private registry.
        // TRIVY_USERNAME / TRIVY_PASSWORD are Trivy's built-in env vars for registry auth.
        withCredentials([string(credentialsId: 'ICR_APIKEY', variable: 'ICR_KEY')]) {
            sh """
                echo "Generating Trivy SBOM from registry..."
                TRIVY_USERNAME=iamapikey TRIVY_PASSWORD="\${ICR_KEY}" \\
                    trivy image ${imageSource} --format spdx-json --output ${trivyFile}
            """
        }
    }
    sh "ls -lh ${trivyFile}"
}

def enrichWithParlay(String trivyFile, String parlayFile) {
    sh """
        echo "Enriching SBOM with Parlay..."
        parlay ecosystems enrich ${trivyFile} > ${parlayFile}
        ls -lh ${parlayFile}
    """
}

def runLicenseScan(String trivyFile, String parlayFile) {
    sh """
        echo "Running license scan..."
        python3 .github/scripts/license_scan.py ${trivyFile} ${parlayFile}
    """
}

def archiveSBOMs(String trivyFile, String parlayFile) {
    archiveArtifacts artifacts: "${trivyFile},${parlayFile}", fingerprint: true
    echo "SBOM artifacts archived"
}

def call(Map config = [:]) {
    def imageSource = config.imageSource ?: error("imageSource is required")
    def trivyFile   = config.trivyFile   ?: 'trivy.json'
    def parlayFile  = config.parlayFile  ?: 'parlay.json'

    installTools()
    generateTrivySBOM(imageSource, trivyFile)
    enrichWithParlay(trivyFile, parlayFile)
    runLicenseScan(trivyFile, parlayFile)
    archiveSBOMs(trivyFile, parlayFile)
}
