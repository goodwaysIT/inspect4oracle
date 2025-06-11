// Language packs
const langMap = {
    'zh': {
        // Text for the index.html page
        'title': 'Oracle Database Inspection',
        'business': 'Business Name',
        'business_placeholder': 'Enter business name',
        'host': 'Host',
        'port': 'Port',
        'service': 'Service Name/SID',
        'username': 'Username',
        'password': 'Password',
        'inspection_items': 'Inspection Items',
        'dbinfo': 'DB Info',
        'storage': 'Storage',
        'params': 'Parameters',
        'backup': 'Backup/Recovery',
        'performance': 'Performance',
        'security': 'Security',
        'objects': 'Objects',
        'sessions': 'Sessions',
        'output_lang': 'Output Language',
        'chinese': 'Chinese',
        'english': 'English',
        'japanese': '日本語',
        'submit': 'Start Inspection',
        'connecting': 'Connecting...',
        'in_progress': 'Inspecting...',
        'generate_report': 'Generating Report...',
        'validate_only': 'Validate Connection Only',
        'validating': 'Validating...',
        'inspection_log': 'Inspection Log',
        
        // Text for the report.html page
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
    },
    'en': {
        // Text for the index.html page
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
        'chinese': 'Chinese',
        'english': 'English',
        'japanese': '日本語',
        'submit': 'Start Inspection',
        'connecting': 'Connecting...',
        'in_progress': 'Inspecting...',
        'generate_report': 'Generating report...',
        'validate_only': 'Validate Only',
        'inspection_log': 'Inspection Log',
        'validating': 'Validating...',
        
        // Text for the report.html page
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
    },
    'jp': {
        // Text for the index.html page
        'title': 'Oracleデータベース検査',
        'business': 'Business Name',
        'business_placeholder': 'Enter business name',
        'host': 'ホスト',
        'port': 'ポート',
        'service': 'Service Name/SID',
        'username': 'Username',
        'password': 'パスワード',
        'inspection_items': '検査項目',
        'dbinfo': '基本情報',
        'storage': 'ストレージ',
        'params': 'パラメータ',
        'backup': 'バックアップ・リカバリ',
        'performance': 'パフォーマンス',
        'security': 'セキュリティ',
        'objects': 'オブジェクト',
        'sessions': 'セッション',
        'output_lang': '出力言語',
        'chinese': '中国語',
        'english': '英語',
        'japanese': '日本語',
        'submit': '検査開始',
        'validate_only': '接続のみ検証',
        'inspection_log': '検査記録',
        'validating': '検証中...',
        'connecting': '接続中...',
        'in_progress': '検査中...',
        'generate_report': 'レポート生成中...',
        
        // Text for the report.html page (can be supplemented later)
        'sidebar_title': 'Oracle 検査レポート',
        'db_info': 'データベース情報',
        'inspection_time': '検査時間:',
        'inspection_modules': '検査モジュール',
        'report_overview': 'レポート概要',
        'report_settings': 'レポート設定',
        'display_style': '表示スタイル',
        'style_default': 'デフォルト',
        'style_light': 'ライト',
        'style_blue': 'フレッシュ',
        'compact_mode': 'コンパクトモード',
        'main_title': 'データベース検査レポート',
        'print_btn': '印刷',
        'export_btn': 'エクスポート',
        'summary': '検査概要',
        'summary_desc': 'このレポートには、次の検査モジュールのデータ分析結果が含まれています：'
    }
};

// 将 langMap 附加到 window 对象，使其成为全局变量
window.langMap = langMap;

// Update page text
function updateTexts(lang) {
    document.documentElement.lang = lang;
    const elements = document.querySelectorAll('[data-lang-key]');
    elements.forEach(el => {
        const key = el.getAttribute('data-lang-key');
        if (langMap[lang] && langMap[lang][key]) {
            if (el.tagName === 'INPUT' && (el.type === 'text' || el.type === 'password')) {
                el.placeholder = langMap[lang][key];
            } else if (el.tagName === 'INPUT' && el.type === 'submit') {
                el.value = langMap[lang][key];
            } else if (el.tagName === 'BUTTON') {
                // Add support for button labels
                el.textContent = langMap[lang][key];
            } else if (el.tagName === 'LABEL' || el.tagName === 'SPAN' || el.tagName === 'DIV' || 
                       el.tagName === 'H1' || el.tagName === 'H2' || el.tagName === 'H3' || 
                       el.tagName === 'H4' || el.tagName === 'H5' || el.tagName === 'H6' || 
                       el.tagName === 'P') {
                // Add support for paragraph tags
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
    updateTexts: updateTexts,  // Export updateTexts function for direct use by other pages
    getCurrentLang: function() { return localStorage.getItem('lang') || 'zh'; }  // Get the current language setting
};
