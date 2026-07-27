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
 */
def call(String buildResult) {
    // Map Jenkins result → GitHub state
    // currentBuild.result is null while the build is still running (counts as SUCCESS)
    def result = buildResult ?: 'SUCCESS'
    def state  = (result == 'SUCCESS') ? 'success' : 'failure'
    def desc   = (result == 'SUCCESS') ? 'Build passed' : 'Build failed'

    // JOB_NAME = "pac/pac-test/pac-go-lint/PR-110"
    // Second-to-last segment is always the Multibranch pipeline name e.g. "pac-go-lint"
    def parts   = env.JOB_NAME?.split('/')
    def jobName = (parts && parts.size() >= 2) ? parts[parts.size() - 2] : (parts ? parts[0] : 'jenkins')
    def context = "jenkins/${jobName}"

    // PR_HEAD_SHA is set explicitly right after checkout scm in each Jenkinsfile
    // via: env.PR_HEAD_SHA = sh(script:'git rev-parse refs/remotes/origin/PR-${CHANGE_ID}', ...)
    // This is the only reliable source — both GIT_COMMIT and CHANGE_SHA contain the
    // ephemeral pr-merge SHA when the pr-merge checkout strategy is used, which does
    // not exist on GitHub and causes a 422 on the statuses API.
    def commitSha = env.PR_HEAD_SHA ?: env.CHANGE_SHA ?: env.GIT_COMMIT
    def repoUrl   = env.GIT_URL ?: env.GIT_URL_1 ?: ''
    def repoName  = repoUrl
        .replaceAll('https://github.com/', '')
        .replaceAll('git@github.com:', '')
        .replaceAll('\\.git$', '')

    try {
        withCredentials([string(credentialsId: 'github-pat', variable: 'GH_TOKEN')]) {
            sh """
                echo "Posting GitHub status: state=${state}, context=${context}, repo=${repoName}, sha=${commitSha}"

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
        // Never fail the build just because the GitHub status post failed.
        // The error body is already printed above by --fail-with-body.
        echo "WARNING: Failed to post GitHub status — ${e.message}"
    }
}
