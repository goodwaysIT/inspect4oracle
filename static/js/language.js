// Language packs
const langMap = {
    'zh': {
        // index.html 页面的文本
        'title': 'Oracle数据库巡检',
        'business': '业务名称',
        'business_placeholder': '请输入业务名称',
        'host': '主机地址',
        'port': '端口',
        'service': '服务名/SID',
        'username': '用户名',
        'password': '密码',
        'inspection_items': '巡检项',
        'dbinfo': '数据库信息',
        'storage': '存储',
        'params': '参数',
        'backup': '备份与恢复',
        'performance': '性能',
        'security': '安全',
        'objects': '对象',
        'sessions': '会话',
        'output_lang': '输出语言',
        'chinese': '中文',
        'english': '英文',
        'japanese': '日语',
        'submit': '开始巡检',
        'connecting': '连接中...',
        'in_progress': '巡检中...',
        'generate_report': '生成报告中...',
        'validate_only': '仅验证连接',
        'validating': '验证中...',
        'inspection_log': '巡检日志',

        // report.html 页面的文本
        'sidebar_title': 'Oracle 巡检报告',
        'db_info': '数据库信息',
        'inspection_time': '巡检时间:',
        'inspection_modules': '巡检模块',
        'report_overview': '报告总览',
        'report_settings': '报告设置',
        'display_style': '显示样式',
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
        'business': 'ビジネス名',
        'business_placeholder': 'ビジネス名を入力',
        'host': 'ホスト',
        'port': 'ポート',
        'service': 'サービス名/SID',
        'username': 'ユーザー名',
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
