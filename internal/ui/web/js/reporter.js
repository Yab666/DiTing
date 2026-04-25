/**
 * Reporter 模块：负责将审计结果导出为指定格式的文件
 */

export const exportReport = (results, format) => {
    if (!results || results.length === 0) return;

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    let content = '';
    let mimeType = '';
    let extension = '';

    if (format === 'csv') {
        // 添加 BOM 头防止 Excel 中文乱码
        content = '\uFEFF';
        content += '风险等级,命中规则,文件路径,行号,内容快照\n';
        results.forEach(item => {
            const sev = item.Severity;
            const rule = item.RuleID;
            const file = item.FilePath;
            const line = item.LineNumber;
            const snippet = '"' + (item.Content || '').replace(/"/g, '""') + '"';
            content += `${sev},${rule},${file},${line},${snippet}\n`;
        });
        mimeType = 'text/csv;charset=utf-8;';
        extension = 'csv';
    } else if (format === 'html') {
        const htmlRows = results.map(item => `
            <tr>
                <td><span class="badge ${item.Severity.toLowerCase()}">${item.Severity}</span></td>
                <td>${item.RuleID}</td>
                <td class="font-mono">${item.FilePath}:${item.LineNumber}</td>
                <td><code>${(item.Content || '').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</code></td>
            </tr>
        `).join('');

        content = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>DiTing 安全审计报告</title>
    <style>
        body { font-family: 'Inter', system-ui, sans-serif; background: #050505; color: #ededed; padding: 40px; margin: 0; }
        .container { max-width: 1200px; margin: 0 auto; background: #0a0a0a; border: 1px solid #262626; border-radius: 12px; overflow: hidden; }
        .header { border-bottom: 1px solid #262626; padding: 30px; background: rgba(16, 185, 129, 0.05); }
        h1 { color: #10b981; margin: 0 0 10px 0; font-size: 24px; display: flex; align-items: center; gap: 10px; }
        .meta { color: #a3a3a3; font-size: 14px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 16px 20px; text-align: left; border-bottom: 1px solid #262626; }
        th { background: #121212; color: #a3a3a3; font-weight: 500; font-size: 13px; text-transform: uppercase; letter-spacing: 1px; }
        td { font-size: 14px; vertical-align: middle; }
        tr:last-child td { border-bottom: none; }
        tr:nth-child(even) { background: rgba(255, 255, 255, 0.01); }
        .font-mono { font-family: 'JetBrains Mono', monospace; font-size: 13px; color: #a3a3a3; }
        code { background: #1a1a1a; padding: 6px 10px; border-radius: 6px; font-family: 'JetBrains Mono', monospace; font-size: 12px; color: #10b981; word-break: break-all; display: inline-block; border: 1px solid #262626; }
        .badge { padding: 4px 10px; border-radius: 6px; font-size: 11px; font-weight: 600; display: inline-block; letter-spacing: 0.5px; }
        .badge.critical { background: rgba(239, 68, 68, 0.1); color: #f87171; border: 1px solid rgba(239,68,68,0.2); }
        .badge.major { background: rgba(245, 158, 11, 0.1); color: #fbbf24; border: 1px solid rgba(245,158,11,0.2); }
        .badge.minor { background: rgba(59, 130, 246, 0.1); color: #60a5fa; border: 1px solid rgba(59,130,246,0.2); }
        .badge.info { background: rgba(255, 255, 255, 0.05); color: #a3a3a3; border: 1px solid rgba(255,255,255,0.1); }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>DiTing Security Audit Report</h1>
            <div class="meta">生成时间: ${new Date().toLocaleString()} | 捕获隐患总数: <span style="color:#fff; font-weight:bold;">${results.length}</span></div>
        </div>
        <table>
            <thead>
                <tr><th>风险等级</th><th>规则标识</th><th>文件位置</th><th>内容快照</th></tr>
            </thead>
            <tbody>
                ${htmlRows}
            </tbody>
        </table>
    </div>
</body>
</html>`;
        mimeType = 'text/html;charset=utf-8;';
        extension = 'html';
    }

    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `diting-report-${timestamp}.${extension}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
};
