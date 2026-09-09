#!/usr/bin/env groovy

/**
 * Post a GitHub commit status back to the PR/branch that triggered this build.
 *
 * Usage (call from any post{} block):
 *   notifyGitHub(currentBuild.result)
 *
 * Required Jenkins credential:
 *   ID : github-pat
 *   Kind: Secret text  (a GitHub PAT with repo:status scope)
 *
 * The context label shown on the GitHub PR is taken from the Jenkins job name
 * so each pipeline (lint, unit-tests, docker-api …) shows as its own check.
 *
 * SHA resolution:
 *   With the pr-merge checkout strategy Jenkins builds an ephemeral local merge
 *   commit (PR branch + main). That SHA never exists on GitHub so GIT_COMMIT,
 *   CHANGE_SHA and scmVars.GIT_COMMIT all return this fake SHA → 422 error.
 *   The GitHub API GET /pulls/:number always returns the real PR head SHA.
 *   Falls back to GIT_COMMIT for non-PR branch builds (main, release-*).
 */
def call(String buildResult) {
    // Map Jenkins result → GitHub state
    // currentBuild.result is null while the build is still running (counts as SUCCESS)
    def result = buildResult ?: 'SUCCESS'
    def state  = (result == 'SUCCESS') ? 'success' : 'failure'
    def desc   = (result == 'SUCCESS') ? 'Build passed' : 'Build failed'

    // Second-to-last segment is always the Multibranch pipeline name e.g. "pac-go-lint"
    def parts   = env.JOB_NAME?.split('/')
    def jobName = (parts && parts.size() >= 2) ? parts[parts.size() - 2] : (parts ? parts[0] : 'jenkins')
    def context = "jenkins/${jobName}"

    def repoUrl  = env.GIT_URL ?: env.GIT_URL_1 ?: ''
    def repoName = repoUrl
        .replaceAll('https://github.com/', '')
        .replaceAll('git@github.com:', '')
        .replaceAll('\\.git$', '')

    // Resolve the real PR head SHA via the GitHub API.
    // GIT_COMMIT, CHANGE_SHA and scmVars.GIT_COMMIT all contain the ephemeral
    // pr-merge commit SHA Jenkins creates locally — that SHA does not exist on
    // GitHub and causes a 422 "No commit found for SHA" from the statuses API.
    // GET /pulls/:number returns head.sha which is always the real PR branch tip.
    // python3 is available on all agent images used in this repo.
    def commitSha
    if (env.CHANGE_ID) {
        try {
            withCredentials([string(credentialsId: 'github-pat', variable: 'GH_TOKEN')]) {
                commitSha = sh(
                    script: """
                        curl -sf -H "Authorization: token \${GH_TOKEN}" \
                            "https://api.github.com/repos/${repoName}/pulls/${env.CHANGE_ID}" \
                        | python3 -c "import sys,json; print(json.load(sys.stdin)['head']['sha'])"
                    """,
                    returnStdout: true
                ).trim()
            }
        } catch (Exception e) {
            echo "WARNING: Could not resolve PR head SHA via GitHub API, falling back to GIT_COMMIT — ${e.message}"
            commitSha = env.GIT_COMMIT
        }
    } else {
        // Non-PR build (main, release-*) — GIT_COMMIT is correct here as there
        // is no merge commit involved.
        commitSha = env.GIT_COMMIT
    }

    echo "notifyGitHub: context=${context} CHANGE_ID=${env.CHANGE_ID} sha=${commitSha}"

    try {
        withCredentials([string(credentialsId: 'github-pat', variable: 'GH_TOKEN')]) {
            sh """
                curl -s --fail-with-body -X POST \\
                    -H "Authorization: token \${GH_TOKEN}" \\
                    -H "Content-Type: application/json" \\
                    "https://api.github.com/repos/${repoName}/statuses/${commitSha}" \\
                    -d '{
                        "state":       "${state}",
                        "target_url":  "${env.BUILD_URL}",
                        "description": "${desc}",
                        "context":     "${context}"
                    }'
            """
        }
    } catch (Exception e) {
        // Never fail the build because a GitHub status post failed.
        // The error body is already printed above by --fail-with-body.
        echo "WARNING: Failed to post GitHub status — ${e.message}"
    }
}
