// Global variables
let autoScroll = true;
let startTime = Date.now();
let logEntries = [];
let serverOnline = true;
let isAuthenticated = false;
let currentUser = null;

const enumValue = (name) => Object.freeze({toString: () => name});

const LOG_SEVERITY = Object.freeze({
    ERROR: enumValue("error"),
    WARNING: enumValue("warning"),
    SUCCESS: enumValue("success"),
    INFO: enumValue("info")
});

function init(){
    checkAuthStatus();
    updateServerInfo();
    setInterval(updateServerInfo, 3000); // Update every 3 seconds

    // Enable enter key support for commands
    document.getElementById('command-input').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            processCommand();
        }
    });

    // Enable filter function
    document.getElementById('log-filter').addEventListener('input', filterLogs);
    addLog('Dashboard initialized');
}

function sendCommand(Command, Parameters = []){
    return fetch('/api/command', {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json'
        },
        credentials: 'same-origin', // Include cookies
        body: JSON.stringify({Command, Parameters})
    });
}

function processCommand(){
    const input = document.getElementById('command-input');
    const command = input.value.trim();
    if (!command) return;
    
    input.value = '';
    let params = parse_cmdline(command);
    const cmd = params.shift();
    if(cmd != "login" && cmd != "change_password"){
        addLog(`> ${command}`, 'info');
    }
    sendCommand(cmd, params)
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            addLog(data.message || 'Command executed successfully', 'success');
        } else {
            addLog(data.message || 'Command failed', 'error');
            if(data.update_user || false){
                checkAuthStatus()
            }
        }
    })
    .catch(error => {
        addLog('Command failed: ' + error.message, 'error');
    });
}

function parse_cmdline(cmdline) {
    var re_next_arg = /^\s*((?:(?:"(?:\\.|[^"])*")|(?:'[^']*')|\\.|\S)+)\s*(.*)$/;
    var next_arg = ['', '', cmdline];
    var args = [];
    while (next_arg = re_next_arg.exec(next_arg[2])) {
        var quoted_arg = next_arg[1];
        var unquoted_arg = "";
        while (quoted_arg.length > 0) {
            if (/^"/.test(quoted_arg)) {
                var quoted_part = /^"((?:\\.|[^"])*)"(.*)$/.exec(quoted_arg);
                unquoted_arg += quoted_part[1].replace(/\\(.)/g, "$1");
                quoted_arg = quoted_part[2];
            } else if (/^'/.test(quoted_arg)) {
                var quoted_part = /^'([^']*)'(.*)$/.exec(quoted_arg);
                unquoted_arg += quoted_part[1];
                quoted_arg = quoted_part[2];
            } else if (/^\\/.test(quoted_arg)) {
                unquoted_arg += quoted_arg[1];
                quoted_arg = quoted_arg.substring(2);
            } else {
                unquoted_arg += quoted_arg[0];
                quoted_arg = quoted_arg.substring(1);
            }
        }
        args[args.length] = unquoted_arg;
    }
    return args;
}

function sendActivitie(Activity){
    return fetch('/api/activities', {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json'
        },
        credentials: 'same-origin', // Include cookies
        body: JSON.stringify({activity: Activity})
    }).then(data => data.json()).then(data =>{
        addLog(data.message, data.success == "success" ? LOG_SEVERITY.SUCCESS : LOG_SEVERITY.ERROR)
    })
}

function addLog(message, type = LOG_SEVERITY.INFO) {
    const timestamp = new Date().toLocaleTimeString();
    const logEntry = {
        timestamp,
        message,
        type,
        id: Date.now() + Math.random()
    };
    logEntries.push(logEntry);

    const consoleOutput = document.getElementById('console-output');
    if (!consoleOutput) {
        console.error('Console output element not found');
        return;
    }

    const logElement = document.createElement('div');
    logElement.className = `log-entry log-${type} font-mono`; // monospaced for CMD look

    // Build inner content
    let content = `<span class="log-timestamp">[${timestamp}]</span><br>`;
    if (Array.isArray(message)) {
        content += message.map(line => `> ${line}`).join("<br>");
    } else {
        content += `> ${message}`;
    }

    logElement.innerHTML = content;
    consoleOutput.appendChild(logElement);

    if (autoScroll) {
        consoleOutput.scrollTop = consoleOutput.scrollHeight;
    }
}
function checkAuthStatus() {
    // Try to get session from cookie or make a whoami request
    sendCommand('whoami')
        .then(response => response.json())
        .then(data => {
            console.log('Auth check response data:', data);
            if(data.status == "success"){
                isAuthenticated = true;
                currentUser = data.username;
                updateAuthUI();
                addLog(`Already logged in as: ${currentUser}`, 'success');
            }else{
                // Not authenticated
                isAuthenticated = false;
                currentUser = null;
                updateAuthUI();
                addLog('Please login to access server features', 'warning');
            }
        })
        .catch((error) => {
            console.log('Auth check failed:', error);
            isAuthenticated = false;
            currentUser = null;
            updateAuthUI();
            addLog('Please login to access server features', 'warning');
        });
}

// Update UI based on authentication status
function updateAuthUI() {
    let authStatus = document.getElementById('auth-status');
    const serverControls = document.getElementById('server-controls');
    
    if (!authStatus) {
        // Create auth status element if it doesn't exist
        const statusContainer = document.querySelector('.server-info') || document.body;
        const authDiv = document.createElement('div');
        authDiv.id = 'auth-status';
        authDiv.className = 'auth-status';
        statusContainer.appendChild(authDiv);
        authStatus = authDiv;
    }
    if (isAuthenticated) {
        authStatus.innerHTML = `
            <span class="auth-indicator online"></span>
            <span>Logged in as: ${currentUser}</span>
            <button onclick="logout()" class="btn-logout">Logout</button>
        `;
        // Enable server controls
        if (serverControls) {
            serverControls.style.opacity = '1';
            serverControls.style.pointerEvents = 'auto';
        }
    } else {
        authStatus.innerHTML = `
            <span class="auth-indicator offline"></span>
            <span>Not authenticated - Please login</span>
        `;
        
        // Disable server controls
        if (serverControls) {
            serverControls.style.opacity = '0.5';
            serverControls.style.pointerEvents = 'none';
        }
    }
}

function updateServerInfo() {
    fetch('/api/status', {
        credentials: 'same-origin' // Include cookies
    })
    .then(response => response.json())
    .then(data => {
        serverOnline = true;
        updateServerStatus();
        if(data.status != "success"){
            isAuthenticated = false;
            currentUser = null;
            updateAuthUI();
            return
        }
        
        // Update server info display
        const portEl = document.getElementById('server-port');
        const uptimeEl = document.getElementById('uptime');
        const cpuEl = document.getElementById('cpu-usage');
        const memoryEl = document.getElementById('memory-usage');
        const connectionsEl = document.getElementById('connections');
        const lastUpdateEl = document.getElementById('last-update');
        
        if (portEl) portEl.textContent = data.port || 'N/A';

        if (data.startTime && uptimeEl) {
            const startTime = new Date(data.startTime);
            const uptime = Date.now() - startTime.getTime();
            const hours = Math.floor(uptime / 3600000);
            const minutes = Math.floor((uptime % 3600000) / 60000);
            const seconds = Math.floor((uptime % 60000) / 1000);
            uptimeEl.textContent = 
                `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
        }

        if (cpuEl) {
            cpuEl.textContent = (typeof data.cpu === 'number' ? data.cpu.toFixed(1) + '%' : data.cpu || '0%');
        }
        if (memoryEl) {
            memoryEl.textContent = (data.memory ? Math.round(data.memory / 1024 / 1024) + ' MB' : '0 MB');
        }
        if (connectionsEl) {
            connectionsEl.textContent = data.connections || '0';
        }
        if (lastUpdateEl) {
            lastUpdateEl.textContent = new Date().toLocaleTimeString();
        }
        
        isAuthenticated = true;
        currentUser = data.username;
        updateAuthUI();
    })
    .catch(error => {
        console.error('Failed to fetch server status:', error);
        serverOnline = false;
        updateServerStatus();
        
        // If we can't reach the server, we're effectively not authenticated
        if (isAuthenticated) {
            isAuthenticated = false;
            currentUser = null;
            updateAuthUI();
        }
    });
}

function updateServerStatus() {
    const statusElement = document.getElementById('server-status');
    const indicatorElement = document.getElementById('status-indicator');
    
    if (statusElement) {
        statusElement.textContent = serverOnline ? 'Online' : 'Offline';
    }
    if (indicatorElement) {
        indicatorElement.className = serverOnline ? 'status-indicator status-online' : 'status-indicator status-offline';
    }
}

function clearConsole() {
    const consoleOutput = document.getElementById('console-output');
    if (consoleOutput) {
        consoleOutput.innerHTML = '';
    }
    logEntries = [];
    addLog('Console cleared', 'info');
}

function toggleAutoScroll() {
    autoScroll = !autoScroll;
    const btn = document.getElementById('autoscroll-btn');
    if (btn) {
        btn.textContent = `Auto-scroll: ${autoScroll ? 'ON' : 'OFF'}`;
    }
    showNotification(`Auto-scroll ${autoScroll ? 'enabled' : 'disabled'}`, 'info');
}

function filterLogs() {
    const filter = document.getElementById('log-filter').value.toLowerCase();
    const logElements = document.querySelectorAll('.log-entry');
    
    logElements.forEach(element => {
        const text = element.textContent.toLowerCase();
        element.style.display = text.includes(filter) ? 'block' : 'none';
    });
}

function showNotification(message, type = 'info') {
    let notification = document.getElementById('notification');
    let content = document.getElementById('notification-content');
    
    if (!notification || !content) {
        // Create notification elements if they don't exist
        notification = document.createElement('div');
        notification.id = 'notification';
        notification.className = 'notification';
        
        content = document.createElement('div');
        content.id = 'notification-content';
        
        notification.appendChild(content);
        document.body.appendChild(notification);
    }
    
    content.textContent = message;
    notification.className = `notification ${type} show`;
    
    setTimeout(() => {
        notification.classList.remove('show');
    }, 3000);
}

function quickLogin() {
    const username = prompt('Username:');
    if (!username) return;
    
    const password = prompt('Password:');
    if (!password) return;
    
    processCommand(`login ${username} ${password}`);
}

function logout(){
    sendCommand("logout")
}

// Initialize when page loads
window.addEventListener('DOMContentLoaded', init);

// Export functions for global access
window.quickLogin = quickLogin;
window.logout = logout;
window.serverAction = sendActivitie;
window.toggleAutoScroll = toggleAutoScroll;
window.clearConsole = clearConsole;