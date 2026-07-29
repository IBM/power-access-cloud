#!/usr/bin/env groovy

/**
 * Clean up all workspace directories left behind by the pr-merge checkout strategy.
 *
 * PR merge jobs cannot use lightweight checkout so Jenkins leaves behind
 * workspace@script, workspace@libs and workspace@tmp directories on the
 * server node for every build. These accumulate and consume significant disk.
 *
 * Usage — add to every Jenkinsfile post { always } block:
 *   post { always { pipelineTeardown() } }
 *
 * Requires one executor reserved on the built-in node so the
 * node('built-in') block can run.
 *
 * Reference: http://ibm.biz/taas_jenkins_disk_management
 */
def call() {
    cleanWs()
    dir("${env.WORKSPACE}@tmp") {
        deleteDir()
    }
    node('built-in') {
        dir("${env.WORKSPACE}@libs") {
            deleteDir()
        }
        dir("${env.WORKSPACE}@script") {
            deleteDir()
        }
        dir("${env.WORKSPACE}@script@tmp") {
            deleteDir()
        }
    }
}
