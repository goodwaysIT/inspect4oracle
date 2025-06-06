// 语言包
const langMap = {
    'zh': {
        // index.html 页面文本
        'title': 'Oracle数据库巡检',
        'business': '数据库业务名称',
        'business_placeholder': '请输入业务名称',
        'host': '地址',
        'port': '端口',
        'service': '服务名/SID',
        'username': '用户名',
        'password': '密码',
        'inspection_items': '巡检项',
        'dbinfo': '基本信息',
        'storage': '存储',
        'params': '参数',
        'backup': '备份恢复',
        'performance': '性能',
        'security': '安全',
        'objects': '对象',
        'sessions': '会话',
        'output_lang': '输出语言',
        'chinese': '中文',
        'english': 'English',
        'submit': '开始巡检',
        'connecting': '正在连接...',
        'in_progress': '正在巡检...',
        'generate_report': '生成报告中...',
        'validate_only': '仅验证连接',
        'inspection_log': '检查记录',
        
        // report.html 页面文本
        'sidebar_title': 'Oracle 巡检报告',
        'db_info': '数据库信息',
        'inspection_time': '巡检时间:',
        'inspection_modules': '巡检模块',
        'report_overview': '报告总览',
        'report_settings': '报告设置',
        'display_style': '显示风格',
        'style_default': '默认',
        'style_light': '浅色',
        'style_blue': '清新',
        'compact_mode': '紧凑模式',
        'main_title': '数据库巡检报告',
        'print_btn': '打印',
        'export_btn': '导出',
        'summary': '巡检概要',
        'summary_desc': '本报告包含以下巡检模块的数据分析结果：'
    },
    'en': {
        // index.html 页面文本
        'title': 'Oracle Database Inspection',
        'business': 'Business Name',
        'business_placeholder': 'Enter business name',
        'host': 'Host',
        'port': 'Port',
        'service': 'Service/SID',
        'username': 'Username',
        'password': 'Password',
        'inspection_items': 'Inspection Items',
        'dbinfo': 'Database Info',
        'storage': 'Storage',
        'params': 'Parameters',
        'backup': 'Backup',
        'performance': 'Performance',
        'security': 'Security',
        'objects': 'Objects',
        'sessions': 'Sessions',
        'output_lang': 'Output Language',
        'chinese': '中文',
        'english': 'English',
        'submit': 'Start Inspection',
        'connecting': 'Connecting...',
        'in_progress': 'Inspecting...',
        'generate_report': 'Generating report...',
        'validate_only': 'Validate Only',
        'inspection_log': 'Inspection Log',
        
        // report.html 页面文本
        'sidebar_title': 'Oracle Inspection Report',
        'db_info': 'Database Info',
        'inspection_time': 'Inspection Time:',
        'inspection_modules': 'Inspection Modules',
        'report_overview': 'Report Overview',
        'report_settings': 'Report Settings',
        'display_style': 'Display Style',
        'style_default': 'Default',
        'style_light': 'Light',
        'style_blue': 'Fresh',
        'compact_mode': 'Compact Mode',
        'main_title': 'Database Inspection Report',
        'print_btn': 'Print',
        'export_btn': 'Export',
        'summary': 'Inspection Summary',
        'summary_desc': 'This report contains data analysis results for the following inspection modules:'
    }
};

// 更新页面文本
function updateTexts(lang) {
    const elements = document.querySelectorAll('[data-lang-key]');
    elements.forEach(el => {
        const key = el.getAttribute('data-lang-key');
        if (langMap[lang] && langMap[lang][key]) {
            if (el.tagName === 'INPUT' && (el.type === 'text' || el.type === 'password')) {
                el.placeholder = langMap[lang][key];
            } else if (el.tagName === 'INPUT' && el.type === 'submit') {
                el.value = langMap[lang][key];
            } else if (el.tagName === 'BUTTON') {
                // 添加对按钮标签的支持
                el.textContent = langMap[lang][key];
            } else if (el.tagName === 'LABEL' || el.tagName === 'SPAN' || el.tagName === 'DIV' || 
                       el.tagName === 'H1' || el.tagName === 'H2' || el.tagName === 'H3' || 
                       el.tagName === 'H4' || el.tagName === 'H5' || el.tagName === 'H6' || 
                       el.tagName === 'P') {
                // 添加对段落标签的支持
                el.textContent = langMap[lang][key];
                // console.log(`Updated ${el.tagName} element with key '${key}' to: '${langMap[lang][key]}'`);
            }
        }
    });
}

// 初始化语言
function syncLangField() {
    var lang = localStorage.getItem('lang') || 'zh';
    var langInput = document.getElementById('lang');
    if (langInput) langInput.value = lang;
}

function initLanguage() {
    const defaultLang = localStorage.getItem('lang') || 'zh';
    updateTexts(defaultLang);
    syncLangField();
    // 语言切换事件
    document.querySelectorAll('input[name="lang"]').forEach(radio => {
        radio.addEventListener('change', (e) => {
            localStorage.setItem('lang', e.target.value);
            updateTexts(e.target.value);
            syncLangField();
        });
    });
    // 页面加载时同步一次
    document.addEventListener('DOMContentLoaded', syncLangField);
}

// 导出函数
window.languageModule = {
    init: initLanguage,
    updateTexts: updateTexts,  // 导出更新文本函数供其他页面直接调用
    getCurrentLang: function() { return localStorage.getItem('lang') || 'zh'; }  // 获取当前语言设置
};
