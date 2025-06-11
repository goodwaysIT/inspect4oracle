// 表单处理模块
const formHandler = {
    // Show message box
    showMessage(message, type = 'info') {
        // Create message box container
        let messageBox = document.getElementById('message-box');
        
        if (!messageBox) {
            messageBox = document.createElement('div');
            messageBox.id = 'message-box';
            messageBox.style.position = 'fixed';
            messageBox.style.bottom = '20px'; // 修改到屏幕底部
            messageBox.style.right = '20px';
            messageBox.style.zIndex = '9999';
            document.body.appendChild(messageBox);
        }
        
        // Create message element
        const messageElement = document.createElement('div');
        messageElement.className = `alert alert-${type} alert-dismissible fade show`;
        messageElement.role = 'alert';
        messageElement.style.minWidth = '300px';
        messageElement.style.marginBottom = '10px';
        
        // Add message content
        messageElement.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
        `;
        
        // Add to message box
        messageBox.appendChild(messageElement);
        
        // Disappears automatically after 5 seconds
        setTimeout(() => {
            messageElement.classList.remove('show');
            setTimeout(() => messageElement.remove(), 150);
        }, 5000);
    },
    
    // 初始化表单
    init() {
        // console.log('[form-handler.js] DEBUG: formHandler.init() called.');
        // Get form elements - IDs must exactly match those in the HTML
        this.form = document.getElementById('inspection-form'); // 更正为 'inspection-form'
        // console.log('[form-handler.js] DEBUG: this.form element:', this.form);
        this.logContent = document.getElementById('inspection-log-content'); // Ensure logs are output to the correct card
        
        // 如果没有表单（可能在报告页面），则返回
        if (!this.form) {
            // console.log('[form-handler.js] DEBUG: Form element with ID \'inspection-form\' NOT found. Returning from init.');
            return;
        }
        // console.log('[form-handler.js] DEBUG: Form element found. Adding submit event listener.');
        this.form.addEventListener('submit', this.handleFormSubmit.bind(this));
        
        // 初始化密码显示/隐藏切换
        this.initPasswordToggle();
        
        // 初始化折叠面板
        this.initCollapsePanels();
        
        // 初始化验证按钮
        const validateBtn = document.getElementById('validateBtn');
        if (validateBtn) {
            validateBtn.addEventListener('click', this.handleValidateConnection.bind(this));
        }
    },
    
    // 处理验证连接按钮点击
    async handleValidateConnection(event) {
        const button = event.target;
        const form = document.getElementById('inspection-form');
        const formData = new FormData(form);
        const originalButtonText = button.innerHTML;
        
        try {
            // Disable button to prevent multiple clicks
            button.disabled = true;
            const currentLang = document.querySelector('input[name="lang"]:checked')?.value || 'zh';
            const validatingText = (window.langMap && window.langMap[currentLang]?.validating) || '验证中...';
            button.innerHTML = `<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> ${validatingText}`;
            
            // 验证数据库连接
            const validation = await this.validateConnection(formData);
            
            if (validation.success) {
                this.showMessage('Database connection validation successful!', 'success');
            } else {
                this.showMessage(`Validation failed: ${validation.message}`, 'danger');
            }
            
        } catch (error) {
            // console.error('Validation error:', error);
            this.showMessage(`Error occurred: ${error.message}`, 'danger');
        } finally {
            // Restore button state
            button.disabled = false;
            button.innerHTML = originalButtonText;
        }
    },
    
    // 处理表单提交
    async handleFormSubmit(event) {
        event.preventDefault();
        
        const form = event.target;
        const submitButton = form.querySelector('button[type="submit"]');
        const originalButtonText = submitButton.innerHTML;
        
        const formData = new FormData();
        const formElements = form.elements;
        for (let i = 0; i < formElements.length; i++) {
            const element = formElements[i];
            if (element.name && element.type !== 'checkbox') {
                formData.append(element.name, element.value);
            } else if (element.type === 'checkbox' && element.checked) {
                formData.append('items[]', element.value);
            }
        }
        
        try {
            submitButton.disabled = true;
            const currentLang = document.querySelector('input[name="lang"]:checked')?.value || 'zh';
            const inProgressText = (window.langMap && window.langMap[currentLang]?.in_progress) || '巡检中...';
            submitButton.innerHTML = `<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> ${inProgressText}`;
            
            this.showLogCard();
            this.log('Starting database inspection...', 'info');
            // console.log('[form-handler.js] DEBUG: Before calling submitInspection. FormData entries:');
            for (let pair of formData.entries()) { 
                // console.log('[form-handler.js] DEBUG: FormData: ' + pair[0]+ ', ' + pair[1]); 
            }
            
            const inspectionResult = await this.submitInspection(formData);
            // console.log('[form-handler.js] DEBUG: After calling submitInspection. inspectionResult:', inspectionResult);
            
            if (!inspectionResult.success) {
                this.log(`Inspection failed: ${inspectionResult.message}`, 'error');
                this.showMessage(`Inspection failed: ${inspectionResult.message}`, 'danger');
                return;
            }
            
            this.log('Inspection completed', 'success');
            this.showMessage('Inspection completed, redirecting to report page...', 'success');
            
            if (inspectionResult.reportId) {
                this.log('Redirecting to report page...', 'info');
                // console.log('[form-handler.js] DEBUG: reportId found:', inspectionResult.reportId, '. Attempting redirect.');
                window.location.href = `/report.html?id=${inspectionResult.reportId}`;
            } else {
                // console.log('[form-handler.js] DEBUG: reportId NOT found in inspectionResult:', inspectionResult);
                this.log('Report ID not returned, please check server logs', 'error');
                this.showMessage('Report generation failed: No report ID received', 'danger');
            }
            
        } catch (error) {
            // console.error('[form-handler.js] DEBUG: CATCH BLOCK ERROR in handleFormSubmit:', error);
            this.log(`Error occurred: ${error.message}`, 'error');
            this.showMessage(`Error occurred: ${error.message}`, 'danger');
        } finally {
            submitButton.disabled = false;
            submitButton.innerHTML = originalButtonText;
        }
    },
    
    // 验证数据库连接
    async validateConnection(formData) {
        try {
            // 将 FormData 转换为普通对象
            const formObj = {};
            formData.forEach((value, key) => {
                formObj[key] = value;
            });
            
            // 检查必填字段
            const requiredFields = ['host', 'port', 'service', 'username', 'password'];
            const missingFields = requiredFields.filter(field => !formObj[field]);
            
            if (missingFields.length > 0) {
                return {
                    success: false,
                    message: `Please fill in all required fields: ${missingFields.join(', ')}`
                };
            }

            const response = await fetch('/api/validate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    host: formObj.host,
                    port: formObj.port,
                    service: formObj.service,
                    username: formObj.username,
                    password: formObj.password
                })
            });
            
            const result = await response.json();
            
            if (!response.ok || !result.success) {
                return {
                    success: false,
                    message: result.message || 'Validation failed: unknown error'
                };
            }
            
            return { success: true };
            
        } catch (error) {
            // console.error('Error validating connection:', error);
            return {
                success: false,
                message: error.message || 'Error validating connection'
            };
        }
    },
    
    // Submit inspection request
    async submitInspection(formData) {
        try {
            // 记录表单数据以便调试
            const formObj = {};
            formData.forEach((value, key) => {
                formObj[key] = value;
            });


            // Note: Do not set the Content-Type header; the browser will automatically set the correct multipart/form-data boundary
            const response = await fetch('/api/inspect', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) {
                const error = await response.json().catch(() => ({}));
                return {
                    success: false,
                    message: error.message || 'Failed to submit inspection request'
                };
            }
            
            return await response.json();
            
        } catch (error) {
            // console.error('Error submitting inspection request:', error);
            return {
                success: false,
                message: error.message || 'Error submitting inspection request'
            };
        }
    },
        // 显示报告
    async displayReport(reportId) {
        try {
            const response = await fetch(`/api/report/status?id=${reportId}`);
            const reportData = await response.json();
            
            if (!reportData.success) {
                throw new Error(reportData.message || 'Failed to retrieve report');
            }
            
            // 创建报告容器
            const reportContainer = document.createElement('div');
            reportContainer.className = 'report-container';
            reportContainer.style.padding = '20px';
            reportContainer.style.maxWidth = '1200px';
            reportContainer.style.margin = '0 auto';
            
            // 添加报告标题
            const title = document.createElement('h2');
            title.textContent = '数据库巡检报告';
            title.style.textAlign = 'center';
            title.style.marginBottom = '20px';
            reportContainer.appendChild(title);
            
            // Add database information
            const dbInfo = document.createElement('div');
            dbInfo.className = 'card mb-4';
            dbInfo.innerHTML = `
                <div class="card-header">
                    <h5 class="mb-0">Database Information</h5>
                </div>
                <div class="card-body">
                    <p><strong>Database Name:</strong>${reportData.dbInfo || 'N/A'}</p>
                    <p><strong>生成时间：</strong>${reportData.generatedAt || 'N/A'}</p>
                </div>
            `;
            reportContainer.appendChild(dbInfo);
            
            // Add module information
            if (reportData.modules && reportData.modules.length > 0) {
                const modulesContainer = document.createElement('div');
                reportData.modules.forEach(module => {
                    const moduleElement = document.createElement('div');
                    moduleElement.className = 'card mb-4';
                    moduleElement.innerHTML = `
                        <div class="card-header">
                            <h5 class="mb-0">${module.name || 'Unnamed Module'}</h5>
                        </div>
                        <div class="card-body">
                            ${this.formatModuleContent(module)}
                        </div>
                    `;
                    modulesContainer.appendChild(moduleElement);
                });
                reportContainer.appendChild(modulesContainer);
            }
            
            // 清空日志容器并添加报告
            const logContent = document.getElementById('log-content');
            const logContainer = logContent.parentNode; // Get the parent container
            
            // 显示日志容器
            logContent.classList.remove('d-none');
            
            // 清空并添加报告
            logContent.innerHTML = '';
            logContent.appendChild(reportContainer);
            
            // 滚动到报告顶部
            reportContainer.scrollIntoView({ behavior: 'smooth' });
            
        } catch (error) {
            // console.error('Error getting report:', error);
            this.log(`Error getting report: ${error.message}`, 'error');
            this.showMessage(`Error getting report: ${error.message}`, 'danger');
        }
    },
    
    // 格式化模块内容
    formatModuleContent(module) {
        if (!module) return '';
        
        // 如果有卡片数据，显示卡片
        if (module.cards && module.cards.length > 0) {
            let html = '<div class="row">';
            module.cards.forEach(card => {
                html += `
                    <div class="col-md-6 col-lg-4 mb-3">
                        <div class="card h-100">
                            <div class="card-body">
                                <h6 class="card-subtitle mb-2 text-muted">${card.title || ''}</h6>
                                <p class="card-text">${card.value || ''}</p>
                            </div>
                        </div>
                    </div>
                `;
            });
            html += '</div>';
            return html;
        }
        
        // 如果有表格数据，显示表格
        if (module.table) {
            let html = `<h6>${module.table.name || ''}</h6><div class="table-responsive"><table class="table table-striped table-hover">`;
            
            // 表头
            if (module.table.headers && module.table.headers.length > 0) {
                html += '<thead><tr>';
                module.table.headers.forEach(header => {
                    html += `<th>${header}</th>`;
                });
                html += '</tr></thead>';
            }
            
            // 表格内容
            if (module.table.rows && module.table.rows.length > 0) {
                html += '<tbody>';
                module.table.rows.forEach(row => {
                    html += '<tr>';
                    row.forEach(cell => {
                        html += `<td>${cell || ''}</td>`;
                    });
                    html += '</tr>';
                });
                html += '</tbody>';
            }
            
            html += '</table></div>';
            return html;
        }
        
        // 默认返回空字符串
        return '';
    },
    
    // 检查报告状态
    async checkReportStatus(reportId) {
        try {
            // 直接跳转到报告页面
            window.location.href = `/report.html?id=${reportId}`;
        } catch (error) {
            console.error('Error navigating to report page:', error);
            this.log('Error navigating to report page', 'error');
            this.showMessage('Error navigating to report page', 'danger');
        }
    },
    // 显示日志卡片
    showLogCard() {
        if (this.logContent) {
            this.logContent.classList.remove('d-none');
            this.logContent.scrollIntoView({ behavior: 'smooth' });
        }
    },
    
    // 添加日志
    log(message, type = 'info') {
        if (!this.logContent) return;
        
        const now = new Date();
        const timeString = now.toLocaleTimeString();
        const logEntry = document.createElement('div');
        logEntry.className = `log-entry ${type}`;
        logEntry.innerHTML = `[${timeString}] ${message}`;
        
        this.logContent.appendChild(logEntry);
        this.logContent.scrollTop = this.logContent.scrollHeight;
    },
    
    // 初始化密码显示/隐藏切换
    initPasswordToggle() {
        const passwordInput = document.getElementById('password');
        const existingToggle = document.getElementById('togglePassword');
        
        // 确保密码输入框存在才执行后续操作
        if (!passwordInput) {
            // 当前页面可能是报告页面，没有密码输入框，直接返回
            return;
        }

        // If a password toggle button already exists on the page, use it
        if (existingToggle) {
            existingToggle.addEventListener('click', () => {
                const type = passwordInput.type === 'password' ? 'text' : 'password';
                passwordInput.type = type;
                
                // 切换图标
                const icon = existingToggle.querySelector('i');
                if (icon) {
                    icon.classList.toggle('bi-eye');
                    icon.classList.toggle('bi-eye-slash');
                }
            });
            return;
        }
        

    },
    
    // 初始化折叠面板
    initCollapsePanels() {
        // Use Bootstrap's collapse component
        const collapseElements = document.querySelectorAll('[data-bs-toggle="collapse"]');
        collapseElements.forEach(element => {
            element.addEventListener('click', function() {
                const target = this.getAttribute('data-bs-target');
                const targetElement = document.querySelector(target);
                const icon = this.querySelector('i');
                
                if (targetElement.classList.contains('show')) {
                    icon.classList.remove('bi-chevron-up');
                    icon.classList.add('bi-chevron-down');
                } else {
                    icon.classList.remove('bi-chevron-down');
                    icon.classList.add('bi-chevron-up');
                }
            });
        });
    }
};

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    formHandler.init();
    
    // 初始化语言模块
    if (window.languageModule) {
        languageModule.init();
    }
});
