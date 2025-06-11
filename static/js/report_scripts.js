  // Define CSS variables for different styles
  const styleVariables = {
    default: {
      '--primary-blue': '#1a73e8',
      '--secondary-blue': '#e8f0fe',
      '--text-primary': '#202124',
      '--text-secondary': '#5f6368',
      '--border-color': '#dadce0',
      '--card-shadow': '0 2px 6px rgba(60, 64, 67, 0.15)',
      'body-bg': '#f5f7fa',
      'sidebar-bg': '#ffffff',
      'table-header-bg': '#f5f7fa',
      'table-hover-bg': '#f1f3f4',
      'chart-bg': '#ffffff'
    },
    light: {
      '--primary-blue': '#2196f3',
      '--secondary-blue': '#f0f8ff',
      '--text-primary': '#37474f',
      '--text-secondary': '#607d8b',
      '--border-color': '#e0e0e0',
      '--card-shadow': '0 1px 3px rgba(0, 0, 0, 0.08)',
      'body-bg': '#ffffff',
      'sidebar-bg': '#fafafa',
      'table-header-bg': '#f5f5f5',
      'table-hover-bg': '#f9f9f9',
      'chart-bg': '#ffffff'
    },
    blue: {
      '--primary-blue': '#039be5',
      '--secondary-blue': '#e1f5fe',
      '--text-primary': '#01579b',
      '--text-secondary': '#0277bd',
      '--border-color': '#b3e5fc',
      '--card-shadow': '0 2px 6px rgba(3, 155, 229, 0.15)',
      'body-bg': '#e3f2fd',
      'sidebar-bg': '#ffffff',
      'table-header-bg': '#e1f5fe',
      'table-hover-bg': '#e3f2fd',
      'chart-bg': '#ffffff'
    },
    dark: {
      '--primary-blue': '#8ab4f8', // Light blue for dark backgrounds
      '--secondary-blue': '#303841', // Secondary color for dark backgrounds
      '--text-primary': '#e8eaed',  // Light text
      '--text-secondary': '#9aa0a6', // Secondary light text
      '--border-color': '#5f6368',  // 深色模式边框
      '--card-shadow': '0 2px 6px rgba(0, 0, 0, 0.3)', // 深色卡片阴影
      'body-bg': '#202124', // 深色背景
      'sidebar-bg': '#292a2d', // 深色侧边栏
      'table-header-bg': '#292a2d',
      'table-hover-bg': '#3c4043',
      'chart-bg': '#292a2d'
    },
    highContrast: {
      '--primary-blue': '#0000ff', // 纯蓝
      '--secondary-blue': '#e0e0e0', // 灰色背景
      '--text-primary': '#000000',  // Pure black text
      '--text-secondary': '#333333', // Dark gray text
      '--border-color': '#000000',  // 纯黑边框
      '--card-shadow': 'none', // 无阴影，依赖边框
      'body-bg': '#ffffff', // 纯白背景
      'sidebar-bg': '#f0f0f0',
      'table-header-bg': '#e0e0e0',
      'table-hover-bg': '#cccccc',
      'chart-bg': '#ffffff'
    }
  };

  // 切换报告风格
  function changeReportStyle(styleName) {
    // Update the active state of the buttons
    document.querySelectorAll('.sidebar .btn-group[aria-label="风格选择"] .btn').forEach(btn => {
        btn.classList.remove('active');
    });
    const newActiveButton = document.getElementById(`style-${styleName}`);
    if (newActiveButton) {
        newActiveButton.classList.add('active');
    }

    const root = document.documentElement;
    const selectedStyle = styleVariables[styleName] || styleVariables.default;

    for (const [variable, value] of Object.entries(selectedStyle)) {
      if (variable.startsWith('--')) {
        root.style.setProperty(variable, value);
      } else {
        // Special handling for non-CSS variables (e.g., body background)
        switch(variable) {
          case 'body-bg': document.body.style.backgroundColor = value; break;
          // 可以添加更多特殊处理
        }
      }
    }
    localStorage.setItem('reportStyle', styleName); // Save user's choice

    // 更新图表颜色 (如果Chart.js实例存在)
    if (typeof Chart !== 'undefined' && Chart.instances) {
        Object.values(Chart.instances).forEach(chart => {
            updateChartColors(chart, selectedStyle);
        });
    }
  }

  function updateChartColors(chart, style) {
    const isDark = style['body-bg'] === '#202124'; // 简单判断是否为暗色模式
    const gridColor = style['--border-color'];
    const textColor = style['--text-primary'];
    const titleColor = style['--text-primary'];

    if (chart.options.scales) {
        Object.values(chart.options.scales).forEach(scale => {
            if (scale.grid) {
                scale.grid.color = gridColor;
                scale.grid.borderColor = gridColor;
            }
            if (scale.ticks) {
                scale.ticks.color = textColor;
            }
            if (scale.title) {
                scale.title.color = titleColor;
            }
        });
    }
    if (chart.options.plugins && chart.options.plugins.legend) {
        chart.options.plugins.legend.labels.color = textColor;
    }
    if (chart.options.plugins && chart.options.plugins.title) {
        chart.options.plugins.title.color = titleColor;
    }
    
    // Update dataset colors (example, may require more complex logic)
    chart.data.datasets.forEach(dataset => {
        if (isDark) {
            // Choose different colors for dark mode
            // dataset.borderColor = style['--primary-blue'];
            // dataset.backgroundColor = hexToRgb(style['--primary-blue'], 0.1);
        } else {
            // 恢复默认颜色或亮色模式颜色
            // dataset.borderColor = styleVariables.default['--primary-blue'];
            // dataset.backgroundColor = hexToRgb(styleVariables.default['--primary-blue'], 0.1);
        }
    });

    chart.update('none'); // 'none' 表示不播放动画
  }

  // 颜色工具函数，将HEX颜色转为RGB
  function hexToRgb(hex, alpha = 1) {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);

    if (alpha >= 0 && alpha <= 1) {
      return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    }
    return `rgb(${r}, ${g}, ${b})`;
  }

  // 切换紧凑模式
  function toggleCompactMode(enabled) {
    if (enabled) {
      document.body.classList.add('compact-mode');
    } else {
      document.body.classList.remove('compact-mode');
    }
    localStorage.setItem('compactMode', enabled);
  }

// Helper function to check if an item is an object
function isObject(item) {
  return (item && typeof item === 'object' && !Array.isArray(item));
}

// Deep merge two objects
function deepMerge(target, source) {
  let output = { ...target };
  if (isObject(target) && isObject(source)) {
    Object.keys(source).forEach(key => {
      if (isObject(source[key])) {
        if (!(key in target) || !isObject(target[key])) {
          output[key] = source[key]; // Source's object takes precedence if target's key is not an object or if key not in target
        } else {
          output[key] = deepMerge(target[key], source[key]); // Recurse for nested objects
        }
      } else {
        output[key] = source[key]; // Assign non-object values directly from source
      }
    });
  } else if (isObject(source)) { // If target is not an object but source is, return a clone of source
      return { ...source };
  }
  // If source is not an object, target (or its clone) is returned, or if target also not object, primitive source val.
  return output;
}

// Check and apply saved style on page load
document.addEventListener('DOMContentLoaded', function() {
    // console.log('[report_scripts.js] DOMContentLoaded event fired for styles and charts.');

    // DETAILED ADAPTER CHECK - START
    // console.log('[report_scripts.js] --- BEGIN ADAPTER STATE CHECK ---');
    // if (typeof window.Chart === 'function') {
    //     console.log('[report_scripts.js] window.Chart object found. Version:', window.Chart.version);
    //     console.log('[report_scripts.js] typeof window.Chart.registry:', typeof window.Chart.registry);
    //     console.log('[report_scripts.js] window.Chart.registry object:', window.Chart.registry);

    //     if (window.Chart.registry && typeof window.Chart.registry === 'object') { // Check if registry itself is an object
    //         console.log('[report_scripts.js] Chart.registry IS an object.');
    //         console.log('[report_scripts.js] typeof window.Chart.registry.adapters:', typeof window.Chart.registry.adapters);
    //         console.log('[report_scripts.js] window.Chart.registry.adapters object:', window.Chart.registry.adapters);

    //         if (window.Chart.registry.adapters && typeof window.Chart.registry.adapters === 'object') {
    //             console.log('[report_scripts.js] Chart.registry.adapters IS an object.');
    //             const dateAdapter = window.Chart.registry.adapters._date; // In Chart.js v3/v4, _date is the key for the date adapter
    //             console.log('[report_scripts.js] Chart.registry.adapters._date (the registered date adapter):', dateAdapter);

    //             if (dateAdapter) {
    //                 console.log('[report_scripts.js] Date adapter constructor name:', dateAdapter.name); 
    //                 console.log('[report_scripts.js] typeof dateAdapter.formats:', typeof dateAdapter.formats);
    //                 if (typeof dateAdapter.formats === 'function') {
    //                     try {
    //                         console.log('[report_scripts.js] dateAdapter.formats() result:', dateAdapter.formats());
    //                     } catch (e) {
    //                         console.error('[report_scripts.js] Error calling dateAdapter.formats():', e.toString());
    //                     }
    //                 } else {
    //                     console.warn('[report_scripts.js] dateAdapter.formats is NOT a function.');
    //                 }
    //                 const expectedMethods = ['parse', 'format', 'add', 'diff', 'startOf', 'endOf'];
    //                 expectedMethods.forEach(method => {
    //                     if (dateAdapter[method] && typeof dateAdapter[method] === 'function') {
    //                         console.log(`[report_scripts.js] dateAdapter.${method} is a function.`);
    //                     } else {
    //                         console.warn(`[report_scripts.js] dateAdapter.${method} is NOT a function or is missing.`);
    //                     }
    //                 });
    //             } else {
    //                 console.error('[report_scripts.js] Chart.registry.adapters._date is undefined or null. This means the date-fns adapter did not register correctly.');
    //             }
    //         } else {
    //             console.error('[report_scripts.js] Chart.registry.adapters is NOT an object or not found. This is unexpected if Chart.js loaded correctly.');
    //         }
    //     } else {
    //         console.error('[report_scripts.js] Chart.registry is NOT an object or not found. This is highly unexpected if Chart.js loaded correctly.');
    //     }
    // } else {
    //     console.error('[report_scripts.js] window.Chart object NOT found. Chart.js library may not have loaded.');
    // }
    // console.log('[report_scripts.js] --- END ADAPTER STATE CHECK ---');
    // DETAILED ADAPTER CHECK - END

  // Get the last selected style from local storage
  const savedStyle = localStorage.getItem('reportStyle');
  let activeStyle = 'default';
  if (savedStyle && styleVariables[savedStyle]) {
    activeStyle = savedStyle;
  }
  changeReportStyle(activeStyle);

  // Update the active state of the style buttons
  document.querySelectorAll('.sidebar .btn-group[aria-label="风格选择"] .btn').forEach(btn => {
    btn.classList.remove('active');
  });
  const activeButton = document.getElementById(`style-${activeStyle}`);
  if (activeButton) {
    activeButton.classList.add('active');
  }

  const savedCompactMode = localStorage.getItem('compactMode');
  const compactModeToggle = document.getElementById('compactModeToggle');
  if (savedCompactMode !== null) {
      const isCompact = savedCompactMode === 'true';
      toggleCompactMode(isCompact);
      if (compactModeToggle) compactModeToggle.checked = isCompact;
  }

  // 初始化图表
  document.querySelectorAll('.chart-canvas').forEach(function(canvasElement) {
    const chartType = canvasElement.dataset.chartType;
    const chartDataStr = canvasElement.dataset.chartData;
    const chartOptionsStr = canvasElement.dataset.chartOptions;

    if (!chartDataStr) {
      // console.error('Chart data not found for canvas:', canvasElement.id);
      return;
    }

    try {
      const chartData = JSON.parse(chartDataStr);
      // Parse specific options provided by the server
      const specificParsedOptions = chartOptionsStr ? JSON.parse(chartOptionsStr) : {};

      // Enhanced universal chart configuration
      const currentStyle = styleVariables[localStorage.getItem('reportStyle') || 'default'];
      const gridColor = currentStyle['--border-color'];
      const textColor = currentStyle['--text-primary'];

      const defaultOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: chartData.datasets && chartData.datasets.length > 1,
            position: 'top',
            labels: { color: textColor }
          },
          tooltip: {
            mode: 'index',
            intersect: false,
            backgroundColor: currentStyle['body-bg'] === '#202124' ? '#3c4043' : '#fff',
            titleColor: textColor,
            bodyColor: textColor,
            borderColor: gridColor,
            borderWidth: 1
          },
          title: { // Default title config: hidden unless specified in specific options
              display: false, 
              text: '',
              color: textColor, 
              font: { size: 14, weight: '500' },
              padding: { top: 10, bottom: 10 }
          }
        },
        scales: {
          x: {
            type: 'time', // Crucial for date adapter
            time: {
              tooltipFormat: 'yyyy-MM-dd HH:mm',
              displayFormats: {
                  millisecond: 'HH:mm:ss.SSS',
                  second: 'HH:mm:ss',
                  minute: 'HH:mm',
                  hour: 'HH:mm', // Changed for better hour display
                  day: 'MM-dd',
                  week: 'MM-dd',
                  month: 'yyyy-MM',
                  quarter: 'yyyy QQ',
                  year: 'yyyy'
              }
            },
            title: { display: true, text: '时间', color: textColor }, // Default X-axis title
            ticks: { color: textColor },
            grid: { color: gridColor, borderColor: gridColor }
          },
          y: {
            beginAtZero: true,
            title: { display: true, text: '值', color: textColor }, // Default Y-axis title
            ticks: { color: textColor },
            grid: { color: gridColor, borderColor: gridColor }
          }
        }
      };

      // Deep merge default options with specific options from the server
      const finalChartOptions = deepMerge(defaultOptions, specificParsedOptions);

      // Initialize chart
      try {
        new Chart(canvasElement, {
          type: chartType,
          data: chartData,
          options: finalChartOptions
        });
      } catch (e) { // Inner catch for new Chart() instantiation errors
        // console.error('Error initializing chart instance:', canvasElement.id, e);
      }
    } catch (e) { // Outer catch for errors during data/options parsing or setup
        // console.error('Error parsing chart data or options, or during option setup for canvas:', canvasElement.id, e);
        // Optional: Display error message on the canvas
        const ctx = canvasElement.getContext('2d');
        if (ctx) {
            ctx.clearRect(0, 0, canvasElement.width, canvasElement.height);
            ctx.fillStyle = 'red'; // Or use a color from styleVariables
            ctx.font = '14px Arial';
            ctx.textAlign = 'center';
            ctx.fillText('Chart loading failed (Data/Config Error)', canvasElement.width / 2, canvasElement.height / 2);
        }
      } // Closes document.querySelectorAll('.chart-canvas').forEach
    }); // Corrected comment: Closes document.querySelectorAll('.chart-canvas').forEach
}); // Closes the first document.addEventListener('DOMContentLoaded' for styles and charts);
// (其他代码 ... )

// Function: Display the section with the specified ID and update the active state of the navigation links
function showSection(sectionId, clickedLink) {
  // 隐藏所有部分
  document.querySelectorAll('.report-section').forEach(function(section) {
    section.style.display = 'none';
  });
  
  // Remove the active class from all links
  document.querySelectorAll('.sidebar .nav-link').forEach(function(link) {
    link.classList.remove('active');
  });
  
  // Activate the clicked link
  if (clickedLink) {
    clickedLink.classList.add('active');
  }
  
  // 如果是"全部"视图，则显示总览部分
  // Note: Ensure you have an element with ID "section-all" for the overview
  const allSection = document.getElementById('section-all'); 
  if (sectionId === 'all' && allSection) {
    allSection.style.display = 'block';
  } else {
    // 否则显示特定模块
    const targetSection = document.getElementById('section-' + sectionId);
    if (targetSection) {
      targetSection.style.display = 'block';
    } else if (allSection) { // 如果目标模块未找到，回退到总览
        // console.warn(`Section with ID 'section-${sectionId}' not found. Displaying 'all' section.`);
        allSection.style.display = 'block';
        // 并且确保 "all" 链接被激活
        const allNavLink = document.querySelector('.sidebar .nav-link[data-section-id="all"]');
        if (allNavLink && clickedLink !== allNavLink) {
            if(clickedLink) clickedLink.classList.remove('active'); // Remove the previously incorrectly activated link
            allNavLink.classList.add('active');
        }
    }
  }
  
  // 更新页面标题 (可选, 如果你有标题元素)
  // const reportTitleElement = document.getElementById('reportTitle');
  // if (reportTitleElement) {
  //   if (sectionId === 'all' && clickedLink) {
  //     reportTitleElement.textContent = clickedLink.textContent.trim() || '巡检总览';
  //   } else if (clickedLink) {
  //     reportTitleElement.textContent = clickedLink.textContent.trim() || '模块详情';
  //   }
  // }

  // Save the current active sectionId to localStorage
  localStorage.setItem('activeSectionId', sectionId);
}

// Set event listeners for navigation links and show the initial section on page load
document.addEventListener('DOMContentLoaded', function() {
  const navLinks = document.querySelectorAll('.sidebar .nav-link[data-section-id]');
  
  navLinks.forEach(function(link) {
    link.addEventListener('click', function(event) {
      event.preventDefault();
      const sectionId = this.dataset.sectionId;
      showSection(sectionId, this);
        showSection(sectionId, this);
      });
    });

    // Check if there is a saved active section in localStorage
    const savedSectionId = localStorage.getItem('activeSectionId');
    let initialSectionId = 'all'; // 默认显示总览
    let initialLink = document.querySelector('.sidebar .nav-link[data-section-id="all"]');

    if (savedSectionId) {
        const savedLink = document.querySelector(`.sidebar .nav-link[data-section-id="${savedSectionId}"]`);
        if (savedLink) {
            initialSectionId = savedSectionId;
            initialLink = savedLink;
        } else {
            // If the saved sectionId is invalid (e.g., module removed), reset it
            localStorage.removeItem('activeSectionId');
        }
    }
    
    if (initialLink) {
        showSection(initialSectionId, initialLink);
    } else if (navLinks.length > 0) {
        // Fallback if 'all' link is not found, show the first available section
        navLinks[0].click();
    }
  });