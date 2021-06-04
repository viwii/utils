package version

import (
    "log"
)

var (
    AppName      string // 应用名称
    AppVersion   string // 应用版本
    BuildVersion string // 编译版本
    CommitTime   string // 版本提交时间
    BuildTime    string // 编译时间
    GitRevision  string // Git版本
    GitBranch    string // Git分支
    GoVersion    string // Golang信息
)

// Version 版本信息
func Version() {
    log.Println("App Name:", AppName)
    log.Println("App Version:", AppVersion)
    log.Println("Build Version:", BuildVersion)
    log.Println("Build Time:", BuildTime)
    log.Println("Git Revision:", GitRevision)
    log.Println("Git Branch:", GitBranch)
    log.Println("Golang Version:", GoVersion)
    log.Println("Commit Time:", CommitTime)
}
