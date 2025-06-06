/**
 * 报告导出模块 - 用于生成并导出完整的数据库巡检报告
 * @author GoodwaysIT
 * @version 1.0.0
 */

const ReportExporter = (function() {

  // --- Start i18n ---
  const translations = {
    en: {
      sidebar_title_export: "Oracle Inspection Report",
      db_info_export: "Database Information",
      inspection_modules_export: "Inspection Modules",
      report_generated_at_export: "Report Generated:",
      generation_time_export: "Generated:",
      footer_copyright_export: `© ${new Date().getFullYear()} GoodwaysIT. All rights reserved.`,
      db_info_not_found_sidebar: '<p style="color:red; font-weight:bold;">Sidebar element (.sidebar) not found on page.</p>',
      db_info_container_not_found_sidebar: '<p style="color:red; font-weight:bold;">Database info container (.db-info-container) not found within sidebar (.sidebar).</p>',
      db_info_not_provided_export: '<p style="color:orange; font-weight:bold;">Database information was not correctly provided during collection.</p>',
      main_report_title_export: 'Oracle Database Inspection Report'
    },
    zh: {
      sidebar_title_export: "Oracle 数据库巡检报告",
      db_info_export: "数据库信息",
      inspection_modules_export: "巡检模块",
      report_generated_at_export: "报告生成于:",
      generation_time_export: "生成时间:",
      footer_copyright_export: `© ${new Date().getFullYear()} GoodwaysIT. 保留所有权利。`,
      db_info_not_found_sidebar: '<p style="color:red; font-weight:bold;">侧边栏元素 (.sidebar) 未在页面上找到。</p>',
      db_info_container_not_found_sidebar: '<p style="color:red; font-weight:bold;">数据库信息容器 (.db-info-container) 未在侧边栏 (.sidebar) 内找到。</p>',
      db_info_not_provided_export: '<p style="color:orange; font-weight:bold;">数据库信息未在收集阶段正确提供。</p>',
      main_report_title_export: 'Oracle 数据库巡检报告'
    }
  };

  function getLang() {
    return document.documentElement.lang || 'zh'; // Default to Chinese if lang attribute is not set
  }

  function translate(key) {
    const lang = getLang();
    const translated = translations[lang]?.[key] || translations.zh[key] || key; // Fallback to Chinese, then key itself
    // console.log(`Translating key: ${key}, lang: ${lang}, result: ${translated}`); // For debugging i18n
    return translated;
  }
  // --- End i18n ---

  // 私有变量和方法
  const DEFAULT_STYLES = {
    primaryColor: '#0d6efd',
    bgColor: '#f8f9fa',
    cardBg: '#ffffff',
    textColor: '#212529',
    borderColor: '#dee2e6',
    headerBg: '#ffffff',
    shadowColor: 'rgba(0, 0, 0, 0.075)'
  };

  /**
   * 获取数据库名称和连接信息
   * @returns {Object} 包含数据库名和连接信息的对象
   */
  function getDatabaseInfo() {
    const dbInfoElem = document.querySelector('.db-info-container .fw-medium');
    const dbConnElem = document.querySelector('.db-info-container .small:nth-child(2)');
    
    let dbName = '数据库巡检';
    let dbConn = 'localhost';
    
    if (dbInfoElem && dbInfoElem.textContent) {
      // 提取数据库名，就是第一个空格前的内容
      dbName = dbInfoElem.textContent.split(' ')[0] || '数据库巡检';
    }
    
    if (dbConnElem && dbConnElem.textContent) {
      // 提取连接信息，去除图标元素
      dbConn = dbConnElem.textContent.trim().replace(/^\s*[^\s]+\s*/, '') || 'localhost';
    }

    return { dbName, dbConn };
  }

  /**
   * 生成带格式的时间戳
   * @returns {string} 格式化的时间戳
   */
  function getFormattedTimestamp() {
    const now = new Date();
    return now.getFullYear() + 
      ('0' + (now.getMonth() + 1)).slice(-2) + 
      ('0' + now.getDate()).slice(-2) + '_' + 
      ('0' + now.getHours()).slice(-2) + 
      ('0' + now.getMinutes()).slice(-2);
  }

  /**
   * 生成有效的文件名
   * @param {Object} dbInfo 数据库信息
   * @param {string} timestamp 时间戳
   * @returns {string} 有效的文件名
   */
  function generateFileName(dbInfo, timestamp) {
    return (dbInfo.dbName + '_' + dbInfo.dbConn + '_' + timestamp + '.html')
      .replace(/[\\/:\*\?\"<>\|]/g, '_')  // 替换Windows不允许的文件名字符
      .replace(/\s+/g, '_');                  // 替换空格
  }

   /**
    * 收集报告数据，并将图表转换为图片
    * @returns {Object} 包含所有报告数据的对象
    */
   function collectReportData() {
    const modules = [];
    const navLinks = document.querySelectorAll('.sidebar .nav-link[data-section-id]');
    navLinks.forEach(link => {
      const moduleId = link.dataset.sectionId;
      if (moduleId) {
        const iconElement = link.querySelector('.bi');
        const iconClass = iconElement ? iconElement.className : '';
        const iconMatch = iconClass.match(/bi-([^\s]*)/);
        const icon = iconMatch ? iconMatch[1] : 'file-earmark-text';
        
        modules.push({
          id: moduleId,
          name: link.textContent.trim(),
          icon: icon
        });
      }
    });
    
    let dbInfoHTML = '';
    const sidebarElement = document.querySelector('.sidebar');
    if (sidebarElement) {
        // console.log('[ReportExporter] Sidebar element found:', sidebarElement);
        const dbInfoCard = sidebarElement.querySelector('.db-info-container');
        if (dbInfoCard) {
            // console.log('[ReportExporter] dbInfoCard found INSIDE sidebar:', dbInfoCard);
            dbInfoHTML = dbInfoCard.outerHTML;
            // console.log('[ReportExporter] dbInfoHTML CAPTURED (full content):', dbInfoHTML);
        } else {
            // console.warn('[ReportExporter] dbInfoCard NOT found with selector ".db-info-container" INSIDE .sidebar');
            dbInfoHTML = translate('db_info_container_not_found_sidebar');
        }
    } else {
        // console.warn('[ReportExporter] Sidebar element (.sidebar) NOT found on page.');
        dbInfoHTML = translate('db_info_not_found_sidebar');
    }

    const moduleContents = {};
    
    function getClonedSectionWithImages(sectionId) {
       const sectionElement = document.getElementById(sectionId);
       if (!sectionElement) return '';

       const clone = sectionElement.cloneNode(true);

       // 将 Canvas 转换为 Image
       const canvasElements = clone.querySelectorAll('canvas.chart-canvas'); // 确保canvas有 'chart-canvas' 类
       canvasElements.forEach(canvas => {
           try {
               // 尝试从 Chart.js 实例获取图表图片
               // Chart.js v3+ stores instances in Chart.instances
               let chartInstance = null;
               if (typeof Chart !== 'undefined' && Chart.instances) {
                   // Find the chart instance by its canvas ID or the canvas element itself
                   // This assumes the original canvas on the page has an ID that Chart.js uses
                   // Or, if Chart.js allows getting chart by canvas element directly
                   const originalCanvas = document.getElementById(canvas.id); // Or find original by other means
                   if (originalCanvas) {
                       chartInstance = Chart.getChart(originalCanvas);
                   }
               }

               if (chartInstance) {
                   const img = document.createElement('img');
                   img.src = chartInstance.toBase64Image();
                   img.style.maxWidth = '100%';
                   img.style.height = 'auto';
                   img.alt = `Chart for ${sectionId}`;
                   canvas.parentNode.replaceChild(img, canvas);
               } else {
                   // console.warn(`Chart instance not found for canvas in section ${sectionId}, id: ${canvas.id}. Canvas will not be exported as image.`);
                   // Optionally, leave a placeholder or remove the canvas
                   const p = document.createElement('p');
                   p.textContent = '[Chart could not be exported as image]';
                   canvas.parentNode.replaceChild(p, canvas);
               }
           } catch (e) {
               // console.error('Error converting canvas to image:', e);
               const p = document.createElement('p');
               p.textContent = '[Error exporting chart as image]';
               if (canvas.parentNode) {
                  canvas.parentNode.replaceChild(p, canvas);
               }
           }
       });

       // 特殊处理总览页面的链接
       if (sectionId === 'section-all') {
           const overviewLinks = clone.querySelectorAll('.main-info-card a[onclick*="showSection"]');
           overviewLinks.forEach(link => {
               const originalOnclick = link.getAttribute('onclick');
               if (originalOnclick) {
                   const match = originalOnclick.match(/showSection\s*\(\s*['"]([^'"]+)['"]\s*,\s*document\.querySelector\(['"]\.nav-link\[data-section-id=['"]([^'"]+)['"]\]['"]\)\s*\)/);
                   if (match && match.length === 3) {
                       const targetModuleId = match[1];
                       // console.log(`[ReportExporter] Rewriting overview link for module: ${targetModuleId}. Original onclick: ${originalOnclick}`);
                       // 修改 onclick 以调用导出页面内的函数
                       link.setAttribute('onclick', `showExportedSection('${targetModuleId}'); return false;`);
                       link.href = `#section-${targetModuleId}`; // 添加锚点
                   }
               }
           });
       }
       return clone.innerHTML;
    }

    // 收集总览部分
    const allSection = document.getElementById('section-all');
    if (allSection) {
      const originalDisplayAll = allSection.style.display;
      allSection.style.display = 'block'; // 确保可见以正确克隆和转换图表
      moduleContents['all'] = getClonedSectionWithImages('section-all');
      allSection.style.display = originalDisplayAll;
    }
    
    // 收集其他模块内容
    modules.forEach(module => {
      if (module.id !== 'all') {
        const sectionElement = document.getElementById('section-' + module.id);
        if (sectionElement) {
          const originalDisplay = sectionElement.style.display;
          sectionElement.style.display = 'block'; // 确保可见
          moduleContents[module.id] = getClonedSectionWithImages('section-' + module.id);
          sectionElement.style.display = originalDisplay;
        }
      }
    });

    return {
      modules,
      dbInfoHTML,
      moduleContents
    };
  }
  /**
   * 内联所有Bootstrap核心样式
   * @returns {string} 内联样式字符串
   */
  function getInlinedBootstrapStyles() {
    return `
    /* Bootstrap 核心样式 */
    :root{--bs-blue:#0d6efd;--bs-indigo:#6610f2;--bs-purple:#6f42c1;--bs-pink:#d63384;--bs-red:#dc3545;--bs-orange:#fd7e14;--bs-yellow:#ffc107;--bs-green:#198754;--bs-teal:#20c997;--bs-cyan:#0dcaf0;--bs-white:#fff;--bs-gray:#6c757d;--bs-gray-dark:#343a40;--bs-gray-100:#f8f9fa;--bs-gray-200:#e9ecef;--bs-gray-300:#dee2e6;--bs-gray-400:#ced4da;--bs-gray-500:#adb5bd;--bs-gray-600:#6c757d;--bs-gray-700:#495057;--bs-gray-800:#343a40;--bs-gray-900:#212529;--bs-primary:#0d6efd;--bs-secondary:#6c757d;--bs-success:#198754;--bs-info:#0dcaf0;--bs-warning:#ffc107;--bs-danger:#dc3545;--bs-light:#f8f9fa;--bs-dark:#212529;--bs-primary-rgb:13,110,253;--bs-secondary-rgb:108,117,125;--bs-success-rgb:25,135,84;--bs-info-rgb:13,202,240;--bs-warning-rgb:255,193,7;--bs-danger-rgb:220,53,69;--bs-light-rgb:248,249,250;--bs-dark-rgb:33,37,41;--bs-white-rgb:255,255,255;--bs-black-rgb:0,0,0;--bs-body-color-rgb:33,37,41;--bs-body-bg-rgb:255,255,255;--bs-font-sans-serif:system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial,"Noto Sans","Liberation Sans",sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji";--bs-font-monospace:SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;--bs-gradient:linear-gradient(180deg,rgba(255,255,255,.15),rgba(255,255,255,0));--bs-body-font-family:var(--bs-font-sans-serif);--bs-body-font-size:1rem;--bs-body-font-weight:400;--bs-body-line-height:1.5;--bs-body-color:#212529;--bs-body-bg:#fff}*,::after,::before{box-sizing:border-box}@media (prefers-reduced-motion:no-preference){:root{scroll-behavior:smooth}}body{margin:0;font-family:var(--bs-body-font-family);font-size:var(--bs-body-font-size);font-weight:var(--bs-body-font-weight);line-height:var(--bs-body-line-height);color:var(--bs-body-color);text-align:var(--bs-body-text-align);background-color:var(--bs-body-bg);-webkit-text-size-adjust:100%;-webkit-tap-highlight-color:transparent}
    h1,h2,h3,h4,h5,h6{margin-top:0;margin-bottom:.5rem;font-weight:500;line-height:1.2}h1{font-size:calc(1.375rem + 1.5vw)}@media (min-width:1200px){h1{font-size:2.5rem}}h2{font-size:calc(1.325rem + .9vw)}@media (min-width:1200px){h2{font-size:2rem}}h3{font-size:calc(1.3rem + .6vw)}@media (min-width:1200px){h3{font-size:1.75rem}}h4{font-size:calc(1.275rem + .3vw)}@media (min-width:1200px){h4{font-size:1.5rem}}h5{font-size:1.25rem}h6{font-size:1rem}p{margin-top:0;margin-bottom:1rem}
    .container,.container-fluid,.container-lg,.container-md,.container-sm,.container-xl,.container-xxl{width:100%;padding-right:var(--bs-gutter-x,.75rem);padding-left:var(--bs-gutter-x,.75rem);margin-right:auto;margin-left:auto}@media (min-width:576px){.container,.container-sm{max-width:540px}}@media (min-width:768px){.container,.container-md,.container-sm{max-width:720px}}@media (min-width:992px){.container,.container-lg,.container-md,.container-sm{max-width:960px}}@media (min-width:1200px){.container,.container-lg,.container-md,.container-sm,.container-xl{max-width:1140px}}@media (min-width:1400px){.container,.container-lg,.container-md,.container-sm,.container-xl,.container-xxl{max-width:1320px}}
    .row{--bs-gutter-x:1.5rem;--bs-gutter-y:0;display:flex;flex-wrap:wrap;margin-top:calc(var(--bs-gutter-y) * -1);margin-right:calc(var(--bs-gutter-x) * -.5);margin-left:calc(var(--bs-gutter-x) * -.5)}.row>*{flex-shrink:0;width:100%;max-width:100%;padding-right:calc(var(--bs-gutter-x) * .5);padding-left:calc(var(--bs-gutter-x) * .5);margin-top:var(--bs-gutter-y)}
    .card{position:relative;display:flex;flex-direction:column;min-width:0;word-wrap:break-word;background-color:#fff;background-clip:border-box;border:1px solid rgba(0,0,0,.125);border-radius:.25rem}.card-body{flex:1 1 auto;padding:1rem 1rem}.card-title{margin-bottom:.5rem}.card-subtitle{margin-top:-.25rem;margin-bottom:0}.card-text:last-child{margin-bottom:0}.card-header{padding:.5rem 1rem;margin-bottom:0;background-color:rgba(0,0,0,.03);border-bottom:1px solid rgba(0,0,0,.125)}
    .table{--bs-table-bg:transparent;--bs-table-accent-bg:transparent;--bs-table-striped-color:#212529;--bs-table-striped-bg:rgba(0,0,0,.05);--bs-table-active-color:#212529;--bs-table-active-bg:rgba(0,0,0,.1);--bs-table-hover-color:#212529;--bs-table-hover-bg:rgba(0,0,0,.075);width:100%;margin-bottom:1rem;color:#212529;vertical-align:top;border-color:#dee2e6}.table>:not(caption)>*>*{padding:.5rem .5rem;background-color:var(--bs-table-bg);border-bottom-width:1px;box-shadow:inset 0 0 0 9999px var(--bs-table-accent-bg)}.table>tbody{vertical-align:inherit}.table>thead{vertical-align:bottom}.table>:not(:first-child){border-top:2px solid currentColor}
    .badge{display:inline-block;padding:.35em .65em;font-size:.75em;font-weight:700;line-height:1;color:#fff;text-align:center;white-space:nowrap;vertical-align:baseline;border-radius:.25rem}.badge:empty{display:none}.btn .badge{position:relative;top:-1px}.bg-primary{background-color:#0d6efd!important}.bg-secondary{background-color:#6c757d!important}.bg-success{background-color:#198754!important}.bg-danger{background-color:#dc3545!important}.bg-warning{background-color:#ffc107!important}.bg-info{background-color:#0dcaf0!important}
    .text-primary{color:#0d6efd!important}.text-secondary{color:#6c757d!important}.text-success{color:#198754!important}.text-danger{color:#dc3545!important}.text-warning{color:#ffc107!important}.text-info{color:#0dcaf0!important}.text-center{text-align:center!important}.fw-bold{font-weight:700!important}.fw-medium{font-weight:500!important}.fs-6{font-size:1rem!important}.mb-0{margin-bottom:0!important}.mb-2{margin-bottom:.5rem!important}.mb-3{margin-bottom:1rem!important}.mb-4{margin-bottom:1.5rem!important}.mt-3{margin-top:1rem!important}.mt-4{margin-top:1.5rem!important}.mt-5{margin-top:3rem!important}.py-3{padding-top:1rem!important;padding-bottom:1rem!important}.pt-3{padding-top:1rem!important}.border-top{border-top:1px solid #dee2e6!important}.small{font-size:.875em}
    `;
  }

  /**
   * 获取自定义样式
   * @param {Object} style 自定义样式对象
   * @param {boolean} isCompactMode 是否启用紧凑模式
   * @returns {string} 自定义样式字符串
   */
  function getCustomStyles(style, isCompactMode) {
    // 基本样式变量
    const styleObj = style || DEFAULT_STYLES;
    
    // 基本样式
    let css = `
      /* 自定义报告样式 */
      :root {
        --primary-color: ${styleObj.primaryColor};
        --bg-color: ${styleObj.bgColor};
        --card-bg: ${styleObj.cardBg};
        --text-color: ${styleObj.textColor};
        --border-color: ${styleObj.borderColor};
        --shadow-color: ${styleObj.shadowColor};
        --header-bg: ${styleObj.headerBg};
      }
      
      body {
        background-color: var(--bg-color);
        color: var(--text-color);
        font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
        font-size: 14px;
        line-height: 1.5;
        margin: 0;
        padding: 20px;
      }
      
      .report-header {
        text-align: center;
        margin-bottom: 30px;
        padding-bottom: 15px;
        border-bottom: 1px solid var(--border-color);
      }
      
      .report-header h1 {
        color: var(--primary-color);
        font-size: 24px;
        margin-bottom: 10px;
      }
      
      .report-header .timestamp {
        color: var(--text-secondary);
        font-size: 14px;
      }
      
      .db-info-section {
        background-color: var(--card-bg);
        border-radius: 4px;
        padding: 15px;
        margin-bottom: 30px;
        box-shadow: 0 1px 3px var(--shadow-color);
      }
      
      .db-info-section h2 {
        font-size: 18px;
        margin-bottom: 15px;
        color: var(--primary-color);
      }
      
      .report-section {
        background-color: var(--card-bg);
        border-radius: 4px;
        padding: 15px;
        margin-bottom: 30px;
        box-shadow: 0 1px 3px var(--shadow-color);
      }
      
      .report-section h2 {
        font-size: 18px;
        margin-bottom: 15px;
        color: var(--primary-color);
        border-bottom: 1px solid var(--border-color);
        padding-bottom: 10px;
      }
      
      /* 表格样式 */
      .table {
        width: 100%;
        margin-bottom: 1rem;
        color: var(--text-color);
        border-collapse: collapse;
      }
      
      .table th,
      .table td {
        padding: 0.75rem;
        vertical-align: top;
        border-top: 1px solid var(--border-color);
      }
      
      .table thead th {
        vertical-align: bottom;
        border-bottom: 2px solid var(--border-color);
        background-color: var(--header-bg);
      }
      
      .table tbody tr:nth-of-type(odd) {
        background-color: rgba(0, 0, 0, 0.02);
      }
      
      /* 卡片样式 */
      .card {
        background-color: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 4px;
        margin-bottom: 15px;
      }
      
      .card-header {
        padding: 0.75rem 1.25rem;
        background-color: var(--header-bg);
        border-bottom: 1px solid var(--border-color);
      }
      
      .card-body {
        padding: 1.25rem;
      }
      
      .card-title {
        margin-bottom: 0.75rem;
        font-size: 16px;
      }
      
      /* 打印优化 */
      @media print {
        body {
          background-color: white;
          color: black;
          padding: 0;
          margin: 0;
        }
        
        .report-section,
        .db-info-section {
          box-shadow: none;
          border: 1px solid #ddd;
          break-inside: avoid;
          page-break-inside: avoid;
        }
      }
    `;
    
    // 紧凑模式样式
    if (isCompactMode) {
      css += `
        /* 紧凑模式样式 */
        .card-body { padding: 0.75rem !important; }
        .container-fluid { padding: 0.5rem !important; }
        .table td, .table th { padding: 0.4rem !important; }
        .mb-3 { margin-bottom: 0.5rem !important; }
        .mb-4 { margin-bottom: 1rem !important; }
        .mt-3, .mt-4 { margin-top: 0.5rem !important; }
        .pt-3, .py-3, .p-3 { padding-top: 0.5rem !important; }
        .pb-3, .py-3, .p-3 { padding-bottom: 0.5rem !important; }
      `;
    }
    
    css += `
      /* Exported report card layout enhancements */
      .export-report-section .row {
        display: flex;
        flex-wrap: wrap;
        margin-right: -15px;
        margin-left: -15px;
      }
      .export-report-section .row > [class*="col-"] {
        padding-right: 15px;
        padding-left: 15px;
        margin-bottom: 1rem; /* Ensure consistent spacing for cards */
      }
      /* Attempt to enforce 3 columns for col-lg-4 cards */
      .export-report-section .row > .col-lg-4 {
        flex: 0 0 33.333333%;
        max-width: 33.333333%;
      }
      /* For medium screens, col-md-6 should be 2 columns */
      @media (max-width: 991.98px) and (min-width: 768px) {
        .export-report-section .row > .col-md-6 {
          flex: 0 0 50%;
          max-width: 50%;
        }
      }
      /* For small screens, typically 1 column */
      @media (max-width: 767.98px) {
        .export-report-section .row > [class*="col-"] {
          flex: 0 0 100%;
          max-width: 100%;
        }
      }

      /* Ensure stretched-link works for overview cards */
      .export-report-section .card.main-info-card {
        position: relative; /* Needed for stretched-link */
      }
      .export-report-section .main-info-card .stretched-link::after {
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        left: 0;
        z-index: 1;
        content: "";
        background-color: rgba(0,0,0,0); /* Ensure it's clickable */
      }
    `;
    css += `
      /* Table Styles */
      .export-report-section table {
        width: 100%;
        margin-bottom: 1rem;
        color: #212529;
        border-collapse: collapse;
        border: 1px solid #dee2e6;
      }
      .export-report-section table th,
      .export-report-section table td {
        padding: 0.75rem;
        vertical-align: top;
        border-top: 1px solid #dee2e6;
        border-right: 1px solid #dee2e6; /* Added right border for all cells */
      }
      .export-report-section table th:first-child,
      .export-report-section table td:first-child {
        border-left: 1px solid #dee2e6; /* Added left border for first cell in a row */
      }
      .export-report-section table thead th {
        vertical-align: bottom;
        border-bottom: 2px solid #dee2e6;
        background-color: #f8f9fa; /* Light grey background for headers */
        font-weight: bold;
        text-align: left; /* Ensure header text is aligned left */
      }
      .export-report-section table tbody + tbody {
        border-top: 2px solid #dee2e6;
      }
      .export-report-section table tbody tr:nth-of-type(odd) {
        background-color: rgba(0, 0, 0, 0.03); /* Subtle striping for odd rows */
      }
      .export-report-section table caption {
        padding-top: 0.75rem;
        padding-bottom: 0.75rem;
        color: #6c757d;
        text-align: left;
        caption-side: bottom;
      }
    `;
    return css;
  }

   /**
    * 生成左侧导航菜单 (包含数据库信息)
    * @param {Array} modules 模块数组
    * @param {string} dbInfoHtml 数据库信息的HTML字符串
    * @returns {string} 导航菜单HTML
    */
   function generateSidebar(modules, dbInfoHtml) {
    let navLinksHtml = modules.map(module => `
      <li class="nav-item">
        <a class="nav-link${module.id === 'all' ? ' active' : ''}" href="#section-${module.id}" onclick="showExportedSection('${module.id}'); return false;">
          <i class="bi bi-${module.icon || 'file-earmark-text'} me-2"></i>
          ${module.name}
        </a>
      </li>
    `).join('');

    return `
      <nav class="col-md-3 col-lg-2 sidebar">
        <div class="d-flex align-items-center p-3 mb-3 border-bottom">
          <i class="bi bi-database-check me-2 text-primary fs-4"></i>
          <span class="fs-5 fw-bold">${translate('sidebar_title_export')}</span>
        </div>
        <div class="px-3 py-3">
          <h6 class="sidebar-heading text-uppercase mb-2">${translate('db_info_export')}</h6>
          ${dbInfoHtml || translate('db_info_not_provided_export')}
        </div>
        <div class="px-3">
          <div class="d-flex align-items-center my-3">
            <i class="bi bi-layers me-2 text-secondary"></i>
            <h6 class="sidebar-heading mb-0 text-uppercase">${translate('inspection_modules_export')}</h6>
          </div>
          <ul class="nav flex-column mb-4">
            ${navLinksHtml}
          </ul>
        </div>
        <div class="mt-auto px-3 py-3 border-top">
          <p class="small text-muted mb-0">${translate('report_generated_at_export')} ${new Date().toLocaleString()}</p>
        </div>
      </nav>
    `;
  }

  /**
   * 生成完整的HTML文档
   * @param {Object} data 报告数据
   * @param {Object} style 样式配置
   * @param {boolean} isCompactMode 是否启用紧凑模式
   * @returns {string} 完整的HTML文档字符串
   */
  function generateHtmlDocument(data, style, isCompactMode) {
    const now = new Date();
    const modules = data.modules;
    
    // 创建模块内容HTML
    let sectionsHtml = '';
    
    // 生成每个模块的内容区域
    Object.keys(data.moduleContents).forEach(moduleId => {
      const isActive = moduleId === 'all' ? ' active' : '';
      sectionsHtml += `
      <div class="export-report-section${isActive}" id="section-${moduleId}">
        ${data.moduleContents[moduleId]}
      </div>
      `;
    });
    
    return `
    <!DOCTYPE html>
    <html lang="zh">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Oracle数据库巡检报告</title>
      <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.2/font/bootstrap-icons.min.css">
      <style>
        ${getInlinedBootstrapStyles()}
        ${getCustomStyles(style, isCompactMode)}
        
        /* 导出报告特有样式 */
        .container-fluid {
          width: 100%;
          padding-right: 15px;
          padding-left: 15px;
          margin-right: auto;
          margin-left: auto;
        }
        
        .row {
          display: flex;
          flex-wrap: wrap;
          margin-right: -15px;
          margin-left: -15px;
        }
        
        .col-md-3 {
          position: relative;
          width: 25%;
          padding-right: 15px;
          padding-left: 15px;
        }
        
        .col-lg-2 {
          position: relative;
          width: 16.666667%;
          padding-right: 15px;
          padding-left: 15px;
        }
        
        .col-md-9 {
          position: relative;
          width: 75%;
          padding-right: 15px;
          padding-left: 15px;
        }
        
        .col-lg-10 {
          position: relative;
          width: 83.333333%;
          padding-right: 15px;
          padding-left: 15px;
        }
        
        .ms-sm-auto {
          margin-left: auto !important;
        }
        
        .sidebar {
          position: fixed;
          top: 0;
          bottom: 0;
          left: 0;
          z-index: 100;
          padding: 0;
          box-shadow: inset -1px 0 0 rgba(0, 0, 0, .1);
          overflow-y: auto;
          background-color: white;
        }
        
        .sidebar-sticky {
          position: sticky;
          top: 0;
          height: calc(100vh - 48px);
          padding-top: .5rem;
          overflow-x: hidden;
          overflow-y: auto;
        }
        
        .sidebar .nav-link {
          font-weight: 500;
          color: #333;
          padding: 0.5rem 1rem;
          border-left: 3px solid transparent;
          text-decoration: none;
          display: block;
        }
        
        .sidebar .nav-link:hover {
          color: var(--primary-color);
          background-color: rgba(0, 0, 0, 0.05);
        }
        
        .sidebar .nav-link.active {
          color: var(--primary-color);
          background-color: rgba(var(--primary-color-rgb), 0.1);
          border-left-color: var(--primary-color);
        }
        
        .sidebar .nav-link .bi {
          margin-right: 4px;
          color: #999;
        }
        
        .sidebar .nav-link.active .bi {
          color: var(--primary-color);
        }
        
        @media (max-width: 767.98px) {
          .sidebar {
            position: static;
            height: auto;
          }
          
          .sidebar-sticky {
            height: auto;
          }
        }
        
        .export-report-section {
          display: none;
        }
        
        .export-report-section.active {
          display: block;
        }
        
        .nav {
          display: flex;
          flex-wrap: wrap;
          padding-left: 0;
          margin-bottom: 0;
          list-style: none;
        }
        
        .nav-item {
          width: 100%;
        }
        
        .flex-column {
          flex-direction: column !important;
        }
        
        /* 图标修复 */
        .bi {
          display: inline-block;
          vertical-align: -0.125em;
        }
      </style>
    </head>
    <body>
      <div class="container-fluid">
        <div class="row">
          <!-- 左侧导航菜单 -->
          ${generateSidebar(modules)}
          
          <!-- 主体内容区 -->
          <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4 py-3">
            <!-- 内容区头部 -->
            <div class="content-header d-flex justify-content-between align-items-center mb-4">
              <div class="d-flex align-items-center">
                <h1>
                  <i class="bi bi-clipboard-data me-2" style="color: var(--primary-color)"></i>
                  ${translate('main_report_title_export')}
                </h1>
              </div>
              <div class="timestamp text-muted">${translate('generation_time_export')} ${now.toLocaleString()}</div>
            </div>
            
            <!-- 模块内容区 -->
            ${sectionsHtml}
            
            <!-- 底部版权信息 -->
            <footer class="col-12 text-center text-muted small mt-5 py-3 border-top">
              <p class="mb-0">${translate('footer_copyright_export')}</p>
            </footer>
          </main>
        </div>
      </div>
      
      <script>
        function showExportedSection(sectionIdToShow) {
          // 隐藏所有模块内容区域
          document.querySelectorAll('.export-report-section').forEach(section => {
            section.style.display = 'none';
          });
          // 显示目标模块内容区域
          const targetSection = document.getElementById('section-' + sectionIdToShow);
          if (targetSection) {
            targetSection.style.display = 'block';
          }
          // 更新侧边栏导航链接的激活状态
          document.querySelectorAll('.sidebar .nav-link').forEach(link => {
            link.classList.remove('active');
            // 检查 href 是否匹配 (更可靠的方式)
            if (link.getAttribute('href') === '#section-' + sectionIdToShow || 
                (link.getAttribute('onclick') && link.getAttribute('onclick').includes("showExportedSection('" + sectionIdToShow + "')"))) {
              link.classList.add('active');
            }
          });
          // 滚动到模块顶部 (可选)
          if (targetSection) {
            // window.scrollTo(0, targetSection.offsetTop - 20); // 减去一些偏移量
          }
        }
        // 页面加载时，默认显示总览 (all)
        document.addEventListener('DOMContentLoaded', function() {
          showExportedSection('all');
        });
      </script>
    </body>
    </html>
    `;
  }

  /**
   * 保存HTML到文件
   * @param {string} html HTML内容
   * @param {string} fileName 文件名
   */
  function saveHtmlToFile(html, fileName) {
    const blob = new Blob([html], { type: 'text/html;charset=utf-8' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(link.href);
  }

  // 公共API
  return {
    /**
     * 导出报告为HTML文件
     */
    exportReport: function() {
      // 获取数据库信息
      const dbInfo = getDatabaseInfo();
      
      // 生成时间戳和文件名
      const timestamp = getFormattedTimestamp();
      const fileName = generateFileName(dbInfo, timestamp);
      
      // 收集报告数据
      const reportData = collectReportData();
      
      // 获取当前应用的样式设置
      const currentStyle = localStorage.getItem('reportStyle') || 'default';
      const style = window.styleVariables ? window.styleVariables[currentStyle] : DEFAULT_STYLES;
      const isCompactMode = window.compactModeStyles ? window.compactModeStyles.enabled : false;
      
      // 生成HTML文档
      const htmlDocument = generateHtmlDocument(reportData, style, isCompactMode);
      
      // 保存文件
      saveHtmlToFile(htmlDocument, fileName);
    }
  };
})();
