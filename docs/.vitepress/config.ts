const repoRoot = 'https://github.com/alibaba/loongsuite-go-agent';
export default {
    lang: 'en-US',
    title: 'Otel',
    description: 'It provides an automatic solution for Golang applications that want to leverage OpenTelemetry to enable effective observability. No code changes are required in the target application, the instrumentation is done at compile time. Simply adding `otel` prefix to `go build` to get started ', 
    ignoreDeadLinks: true,
    base: '/loongsuite-go-agent/',
    themeConfig: {
        logo: '/anim-logo.svg',
        nav: [
            { 
                text: 'Download',
                items: [
                    { text: 'Linux AMD64', link: `${repoRoot}/releases/latest/download/otel-linux-amd64` },
                    { text: 'Linux ARM64', link: `${repoRoot}/releases/latest/download/otel-linux-arm64` },
                    { text: 'MacOS AMD64', link: `${repoRoot}/releases/latest/download/otel-darwin-amd64` },
                    { text: 'MacOS ARM64', link: `${repoRoot}/releases/latest/download/otel-darwin-arm64` },
                    { text: 'Windows AMD64', link: `${repoRoot}/releases/latest/download/otel-windows-amd64.exe` },
                ]
            },
            {
                text: 'Other Agents',
                items: [
                    { text: 'Go', link: 'https://github.com/alibaba/loongsuite-go-agent' },
                    { text: 'Java', link: 'https://github.com/alibaba/loongsuite-java-agent' },
                    { text: 'Python', link: 'https://github.com/alibaba/loongsuite-python-agent' },
                ]
            }
        ],
        socialLinks: [
            { icon: 'github', link: repoRoot },
            { icon: 'alibabacloud', link: 'https://help.aliyun.com/zh/arms/application-monitoring/getting-started/monitoring-the-golang-applications' }
        ],
        editLink: {
            pattern: `${repoRoot}/edit/main/docs/:path`
        },
        sidebar: [
            {
              text: '🌟 User Guide',
              items: [
                { text: 'Overview', link: '/index' },
                { text: 'Advanced Config', link: '/user/config' },
                { text: 'Compilation Time', link: '/user/compilation-time' },
                { text: 'Experimental', link: '/user/experimental-feature' },
                { text: 'Compatibility', link: '/user/compatibility' },
                { text: 'Manual Instrumentation', link: '/user/manual_instrumentation' },
                { text: 'Context Propagation', link: '/user/context-propagation' },
              ]
            },
            {
                text: '🔧 Developer Guide',
                items: [
                    { text: 'Overview', link: '/dev/overview' },
                    { text: 'Register Hook Rule', link: '/dev/register' },
                    { text: 'Write the Hook Code', link: '/dev/hook' },
                    { text: 'Test the Hook Code', link: '/dev/test' },
                    { text: 'Hook Rule Types', link: '/dev/rule_def' },
                ]
              },
            {
                text: '🤠 Hacking Guide',
                items: [
                  { text: 'Overview', link: '/hacking/overview' },
                  { text: 'Preprocess Phase', link: '/hacking/preprocess' },
                  { text: 'Instrument Phase', link: '/hacking/instrument' },
                  { text: 'AST Optimization', link: '/hacking/optimize' },
                  { text: 'Debugging', link: '/hacking/debug' },
                  { text: 'Tool Internal Slides', link: 'https://github.com/alibaba/loongsuite-go-agent/blob/main/docs/otel-alibaba.pdf' },
                ]
            },
            {
                text: '🌐 Community',
                items: [
                    { text: 'DingTalk', link: 'https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11' },
                ]
            },
        ]
    }
}