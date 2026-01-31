package protocol

// getDashboardHTML returns the main dashboard HTML.
func getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mint DevTools Dashboard</title>
    <style>
        :root {
            --bg-primary: #1e1e1e;
            --bg-secondary: #252526;
            --bg-tertiary: #2d2d30;
            --bg-card: #333333;
            --border-color: #3e3e42;
            --text-primary: #cccccc;
            --text-secondary: #858585;
            --text-muted: #5a5a5a;
            --accent-blue: #569cd6;
            --accent-cyan: #4ec9b0;
            --accent-green: #6a9955;
            --accent-yellow: #dcdcaa;
            --accent-orange: #ce9178;
            --accent-red: #f48771;
            --accent-purple: #c586c0;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            overflow: hidden;
        }

        /* Header */
        .header {
            background: var(--bg-secondary);
            border-bottom: 1px solid var(--border-color);
            padding: 0 16px;
            height: 48px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .header-left {
            display: flex;
            align-items: center;
            gap: 16px;
        }

        .header-title {
            font-size: 16px;
            font-weight: 500;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .header-title .icon { font-size: 18px; }

        .status-bar {
            display: flex;
            align-items: center;
            gap: 20px;
            font-size: 13px;
        }

        .status-item {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: var(--accent-cyan);
            animation: pulse 2s infinite;
        }

        .status-dot.disconnected {
            background: var(--accent-red);
            animation: none;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .metric-badge {
            background: var(--bg-tertiary);
            padding: 4px 10px;
            border-radius: 4px;
            font-size: 12px;
        }

        .metric-badge .value {
            color: var(--accent-cyan);
            font-weight: 600;
            font-family: monospace;
        }

        /* Main Layout */
        .main-container {
            display: flex;
            height: calc(100vh - 48px);
        }

        /* Sidebar */
        .sidebar {
            width: 220px;
            background: var(--bg-secondary);
            border-right: 1px solid var(--border-color);
            display: flex;
            flex-direction: column;
        }

        .sidebar-section {
            padding: 12px 8px;
            border-bottom: 1px solid var(--border-color);
        }

        .sidebar-section-title {
            font-size: 11px;
            text-transform: uppercase;
            color: var(--text-secondary);
            padding: 4px 12px;
            letter-spacing: 0.5px;
        }

        .sidebar-item {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 8px 12px;
            cursor: pointer;
            border-radius: 4px;
            font-size: 13px;
            color: var(--text-primary);
            transition: all 0.15s ease;
            margin: 2px 4px;
        }

        .sidebar-item:hover {
            background: var(--bg-tertiary);
        }

        .sidebar-item.active {
            background: var(--accent-blue);
            color: white;
        }

        .sidebar-item .icon {
            font-size: 16px;
            width: 20px;
            text-align: center;
        }

        .sidebar-footer {
            margin-top: auto;
            padding: 12px;
            border-top: 1px solid var(--border-color);
            font-size: 11px;
            color: var(--text-secondary);
        }

        /* Content Area */
        .content {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
        }

        /* Cards */
        .card {
            background: var(--bg-secondary);
            border-radius: 6px;
            padding: 16px;
            margin-bottom: 16px;
            border: 1px solid var(--border-color);
        }

        .card-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 12px;
            padding-bottom: 8px;
            border-bottom: 1px solid var(--border-color);
        }

        .card-title {
            font-size: 14px;
            font-weight: 500;
            color: var(--accent-blue);
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .card-actions {
            display: flex;
            gap: 8px;
        }

        .btn {
            background: var(--bg-tertiary);
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            padding: 6px 12px;
            border-radius: 4px;
            font-size: 12px;
            cursor: pointer;
            transition: all 0.15s ease;
        }

        .btn:hover {
            background: var(--bg-card);
        }

        .btn-primary {
            background: var(--accent-blue);
            border-color: var(--accent-blue);
            color: white;
        }

        .btn-primary:hover {
            background: #4a8cc7;
        }

        .btn-sm {
            padding: 4px 8px;
            font-size: 11px;
        }

        /* Metrics Grid */
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 12px;
            margin-bottom: 16px;
        }

        .metric-card {
            background: var(--bg-tertiary);
            border-radius: 6px;
            padding: 16px;
            border: 1px solid var(--border-color);
        }

        .metric-card-label {
            font-size: 12px;
            color: var(--text-secondary);
            margin-bottom: 4px;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .metric-card-value {
            font-size: 24px;
            font-weight: 600;
            font-family: monospace;
        }

        .metric-card-value.fps { color: var(--accent-cyan); }
        .metric-card-value.time { color: var(--accent-yellow); }
        .metric-card-value.memory { color: var(--accent-purple); }
        .metric-card-value.count { color: var(--accent-blue); }

        .metric-card-delta {
            font-size: 11px;
            margin-top: 4px;
            color: var(--text-secondary);
        }

        .metric-card-delta.positive { color: var(--accent-green); }
        .metric-card-delta.negative { color: var(--accent-red); }

        /* Charts */
        .chart-container {
            height: 120px;
            background: var(--bg-tertiary);
            border-radius: 4px;
            padding: 12px;
            position: relative;
            overflow: hidden;
        }

        .chart-bars {
            display: flex;
            align-items: flex-end;
            gap: 2px;
            height: 100%;
        }

        .chart-bar {
            flex: 1;
            background: var(--accent-blue);
            border-radius: 2px 2px 0 0;
            min-height: 4px;
            transition: height 0.3s ease;
        }

        .chart-bar:hover {
            background: var(--accent-cyan);
        }

        /* Frame List */
        .frame-list {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }

        .frame-item {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 10px 12px;
            background: var(--bg-tertiary);
            border-radius: 4px;
            font-size: 12px;
            cursor: pointer;
            transition: all 0.15s ease;
        }

        .frame-item:hover {
            background: var(--bg-card);
        }

        .frame-id {
            background: var(--accent-blue);
            color: white;
            padding: 2px 8px;
            border-radius: 3px;
            font-family: monospace;
            font-size: 11px;
        }

        .frame-info {
            flex: 1;
            display: flex;
            gap: 16px;
            color: var(--text-secondary);
        }

        .frame-info span {
            display: flex;
            align-items: center;
            gap: 4px;
        }

        .frame-timestamp {
            color: var(--text-muted);
            font-size: 11px;
        }

        /* Component Tree */
        .component-tree {
            font-family: monospace;
            font-size: 13px;
        }

        .tree-node {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 4px 8px;
            cursor: pointer;
            border-radius: 4px;
        }

        .tree-node:hover {
            background: var(--bg-tertiary);
        }

        .tree-node.selected {
            background: var(--accent-blue);
            color: white;
        }

        .tree-children {
            margin-left: 20px;
            border-left: 1px solid var(--border-color);
        }

        .component-type {
            color: var(--accent-cyan);
        }

        .component-id {
            color: var(--text-secondary);
        }

        /* Snapshot Cards */
        .snapshot-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 12px;
        }

        .snapshot-card {
            background: var(--bg-tertiary);
            border-radius: 6px;
            padding: 12px;
            cursor: pointer;
            transition: all 0.15s ease;
            border: 1px solid transparent;
        }

        .snapshot-card:hover {
            border-color: var(--accent-blue);
            transform: translateY(-2px);
        }

        .snapshot-card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }

        .snapshot-frame-id {
            font-weight: 600;
            color: var(--accent-blue);
        }

        .snapshot-badge {
            background: var(--bg-card);
            padding: 2px 8px;
            border-radius: 10px;
            font-size: 11px;
            color: var(--text-secondary);
        }

        .snapshot-info {
            font-size: 11px;
            color: var(--text-secondary);
        }

        /* Diff View */
        .diff-container {
            font-family: monospace;
            font-size: 12px;
        }

        .diff-item {
            display: flex;
            padding: 6px 8px;
            border-radius: 4px;
            margin-bottom: 2px;
        }

        .diff-item.added {
            background: rgba(106, 153, 85, 0.2);
            color: var(--accent-green);
        }

        .diff-item.removed {
            background: rgba(244, 135, 113, 0.2);
            color: var(--accent-red);
        }

        .diff-item.modified {
            background: rgba(220, 220, 170, 0.2);
            color: var(--accent-yellow);
        }

        .diff-icon {
            width: 20px;
        }

        /* Loading State */
        .loading {
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 40px;
            color: var(--text-secondary);
        }

        .loading::after {
            content: '';
            width: 20px;
            height: 20px;
            border: 2px solid var(--border-color);
            border-top-color: var(--accent-blue);
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-left: 12px;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        /* Empty State */
        .empty-state {
            text-align: center;
            padding: 40px;
            color: var(--text-secondary);
        }

        .empty-state .icon {
            font-size: 48px;
            margin-bottom: 12px;
            opacity: 0.5;
        }

        /* Reconnecting Toast */
        .toast {
            position: fixed;
            top: 60px;
            right: 20px;
            background: var(--accent-red);
            color: white;
            padding: 12px 20px;
            border-radius: 4px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
            display: none;
            align-items: center;
            gap: 10px;
            z-index: 1000;
        }

        .toast.show { display: flex; }

        /* Table */
        .table-container {
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 13px;
        }

        th, td {
            padding: 10px 12px;
            text-align: left;
            border-bottom: 1px solid var(--border-color);
        }

        th {
            background: var(--bg-tertiary);
            font-weight: 500;
            color: var(--text-secondary);
            font-size: 12px;
            text-transform: uppercase;
        }

        tr:hover td {
            background: var(--bg-tertiary);
        }

        /* Badge */
        .badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 10px;
            font-size: 11px;
            font-weight: 500;
        }

        .badge-blue { background: rgba(86, 156, 214, 0.2); color: var(--accent-blue); }
        .badge-green { background: rgba(106, 153, 85, 0.2); color: var(--accent-green); }
        .badge-yellow { background: rgba(220, 220, 170, 0.2); color: var(--accent-yellow); }
        .badge-red { background: rgba(244, 135, 113, 0.2); color: var(--accent-red); }
        .badge-purple { background: rgba(197, 134, 192, 0.2); color: var(--accent-purple); }

        /* Two Column Layout */
        .two-column {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 16px;
        }

        /* Detail Panel */
        .detail-panel {
            background: var(--bg-tertiary);
            border-radius: 4px;
            padding: 12px;
            font-size: 12px;
        }

        .detail-row {
            display: flex;
            justify-content: space-between;
            padding: 4px 0;
            border-bottom: 1px solid var(--border-color);
        }

        .detail-row:last-child {
            border-bottom: none;
        }

        .detail-label {
            color: var(--text-secondary);
        }

        .detail-value {
            font-family: monospace;
            color: var(--text-primary);
        }

        /* Search Input */
        .search-box {
            display: flex;
            align-items: center;
            background: var(--bg-tertiary);
            border-radius: 4px;
            padding: 8px 12px;
            margin-bottom: 12px;
        }

        .search-box input {
            flex: 1;
            background: none;
            border: none;
            color: var(--text-primary);
            font-size: 13px;
            outline: none;
        }

        .search-box .icon {
            color: var(--text-secondary);
            margin-right: 8px;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="header-left">
            <div class="header-title">
                <span class="icon">⚡</span>
                <span>Mint DevTools</span>
            </div>
        </div>
        <div class="status-bar">
            <div class="status-item">
                <div class="status-dot" id="wsStatusDot"></div>
                <span id="wsStatusText">连接中...</span>
            </div>
            <div class="metric-badge">
                FPS: <span class="value" id="headerFps">--</span>
            </div>
            <div class="metric-badge">
                帧时间: <span class="value" id="headerFrameTime">--</span> ms
            </div>
            <div class="metric-badge">
                组件: <span class="value" id="headerComponents">--</span>
            </div>
        </div>
    </div>

    <div class="main-container">
        <div class="sidebar">
            <div class="sidebar-section">
                <div class="sidebar-section-title">主视图</div>
                <div class="sidebar-item active" data-view="dashboard">
                    <span class="icon">📊</span>
                    <span>仪表盘</span>
                </div>
                <div class="sidebar-item" data-view="metrics">
                    <span class="icon">📈</span>
                    <span>性能指标</span>
                </div>
                <div class="sidebar-item" data-view="frames">
                    <span class="icon">🎞️</span>
                    <span>帧列表</span>
                </div>
                <div class="sidebar-item" data-view="components">
                    <span class="icon">🧩</span>
                    <span>组件树</span>
                </div>
            </div>
            <div class="sidebar-section">
                <div class="sidebar-section-title">分析工具</div>
                <div class="sidebar-item" data-view="snapshots">
                    <span class="icon">📸</span>
                    <span>快照</span>
                </div>
                <div class="sidebar-item" data-view="diff">
                    <span class="icon">🔍</span>
                    <span>差异对比</span>
                </div>
                <div class="sidebar-item" data-view="report">
                    <span class="icon">📋</span>
                    <span>调试报告</span>
                </div>
            </div>
            <div class="sidebar-footer">
                <div>版本: 1.0.0</div>
                <div id="connectionInfo">未连接</div>
            </div>
        </div>

        <div class="content" id="mainContent">
            <!-- Content will be loaded here -->
        </div>
    </div>

    <div class="toast" id="reconnectingToast">
        <span>⚠️</span>
        <span>连接断开，正在重连...</span>
    </div>

    <script>
        // Application State
        const AppState = {
            ws: null,
            reconnectInterval: null,
            currentView: 'dashboard',
            metrics: null,
            frames: [],
            components: [],
            snapshots: [],
            fpsHistory: new Array(50).fill(0),
            memoryHistory: new Array(50).fill(0),
            selectedFrame: null,
            selectedComponent: null
        };

        // WebSocket Connection
        function connect() {
            AppState.ws = new WebSocket('ws://' + location.host + '/ws');

            AppState.ws.onopen = () => {
                console.log('[DevTools] WebSocket connected');
                updateStatus(true);
                hideToast();
                if (AppState.reconnectInterval) {
                    clearInterval(AppState.reconnectInterval);
                    AppState.reconnectInterval = null;
                }
                // Send handshake
                AppState.ws.send(JSON.stringify({
                    version: '1.0.0',
                    type: 'handshake',
                    id: 'web-dashboard',
                    payload: {
                        client_id: 'web-dashboard',
                        capabilities: ['snapshots', 'metrics', 'frames'],
                        version: '1.0.0',
                        protocol: 'remote'
                    }
                }));
            };

            AppState.ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    handleMessage(msg);
                } catch (err) {
                    console.error('[DevTools] Failed to parse message:', err);
                }
            };

            AppState.ws.onclose = () => {
                console.log('[DevTools] WebSocket disconnected');
                updateStatus(false);
                if (!AppState.reconnectInterval) {
                    AppState.reconnectInterval = setInterval(() => {
                        showToast();
                        connect();
                    }, 2000);
                }
            };

            AppState.ws.onerror = (error) => {
                console.error('[DevTools] WebSocket error:', error);
            };
        }

        function updateStatus(connected) {
            const dot = document.getElementById('wsStatusDot');
            const text = document.getElementById('wsStatusText');
            const connectionInfo = document.getElementById('connectionInfo');

            if (connected) {
                dot.classList.remove('disconnected');
                text.textContent = '已连接';
                connectionInfo.textContent = '已连接到 localhost:8080';
            } else {
                dot.classList.add('disconnected');
                text.textContent = '已断开';
                connectionInfo.textContent = '连接断开';
            }
        }

        function showToast() {
            document.getElementById('reconnectingToast').classList.add('show');
        }

        function hideToast() {
            document.getElementById('reconnectingToast').classList.remove('show');
        }

        // Message Handling
        function handleMessage(msg) {
            console.log('[DevTools] Received:', msg.type);

            switch (msg.type) {
                case 'metrics_updated':
                    handleMetricsUpdate(msg.data);
                    break;
                case 'frame_added':
                    handleFrameAdded(msg.data);
                    break;
                case 'component_updated':
                    handleComponentUpdated(msg.data);
                    break;
                case 'handshake_ack':
                    console.log('[DevTools] Handshake acknowledged');
                    break;
            }
        }

        function handleMetricsUpdate(metrics) {
            AppState.metrics = metrics;

            // Update history for charts
            AppState.fpsHistory.push(metrics.fps || 0);
            AppState.fpsHistory.shift();
            AppState.memoryHistory.push(metrics.memoryUsage || 0);
            AppState.memoryHistory.shift();

            // Update header
            document.getElementById('headerFps').textContent = metrics.fps?.toFixed(0) || '--';
            document.getElementById('headerFrameTime').textContent = metrics.frameTime || '--';
            document.getElementById('headerComponents').textContent = metrics.componentCount || '--';

            // Update current view if needed
            if (AppState.currentView === 'dashboard' || AppState.currentView === 'metrics') {
                updateMetricsDisplay();
            }
        }

        function handleFrameAdded(frame) {
            AppState.frames.unshift(frame);
            if (AppState.frames.length > 100) {
                AppState.frames.pop();
            }

            if (AppState.currentView === 'dashboard' || AppState.currentView === 'frames') {
                updateFramesDisplay();
            }
        }

        function handleComponentUpdated(comp) {
            AppState.components[comp.id] = comp;

            if (AppState.currentView === 'dashboard' || AppState.currentView === 'components') {
                updateComponentsDisplay();
            }
        }

        // View Rendering
        function loadView(view) {
            AppState.currentView = view;

            // Update sidebar active state
            document.querySelectorAll('.sidebar-item').forEach(el => {
                el.classList.toggle('active', el.dataset.view === view);
            });

            const content = document.getElementById('mainContent');

            switch (view) {
                case 'dashboard':
                    renderDashboard();
                    break;
                case 'metrics':
                    content.innerHTML = '<div class="loading">加载性能指标...</div>';
                    fetchAndRender('/api/metrics', renderMetrics);
                    break;
                case 'frames':
                    content.innerHTML = '<div class="loading">加载帧列表...</div>';
                    fetchAndRender('/api/frames', renderFrames);
                    break;
                case 'components':
                    content.innerHTML = '<div class="loading">加载组件列表...</div>';
                    fetchAndRender('/api/components', renderComponents);
                    break;
                case 'snapshots':
                    content.innerHTML = '<div class="loading">加载快照列表...</div>';
                    fetchAndRender('/api/snapshots', renderSnapshots);
                    break;
                case 'diff':
                    renderDiff();
                    break;
                case 'report':
                    content.innerHTML = '<div class="loading">生成报告...</div>';
                    fetchAndRender('/api/report', renderReport);
                    break;
                default:
                    content.innerHTML = '<div class="empty-state"><div class="icon">❓</div><p>未知视图: ' + view + '</p></div>';
            }
        }

        function fetchAndRender(url, renderer) {
            fetch(url)
                .then(r => {
                    if (!r.ok) throw new Error('HTTP ' + r.status);
                    return r.json();
                })
                .then(data => renderer(data))
                .catch(err => {
                    document.getElementById('mainContent').innerHTML =
                        '<div class="empty-state"><div class="icon">⚠️</div><p>加载失败: ' + err.message + '</p></div>';
                });
        }

        function renderDashboard() {
            const content = document.getElementById('mainContent');
            content.innerHTML =
                '<div class="metrics-grid">' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">⚡</span> FPS</div>' +
                        '<div class="metric-card-value fps" id="dashFps">--</div>' +
                        '<div class="metric-card-delta" id="dashFpsDelta"></div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">⏱️</span> 帧时间</div>' +
                        '<div class="metric-card-value time" id="dashFrameTime">--</div>' +
                        '<div class="metric-card-delta">毫秒/帧</div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">💾</span> 内存</div>' +
                        '<div class="metric-card-value memory" id="dashMemory">--</div>' +
                        '<div class="metric-card-delta" id="dashMemoryDelta"></div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">🧩</span> 组件</div>' +
                        '<div class="metric-card-value count" id="dashComponents">--</div>' +
                        '<div class="metric-card-delta" id="dashComponentsDelta"></div>' +
                    '</div>' +
                '</div>' +
                '<div class="two-column">' +
                    '<div class="card">' +
                        '<div class="card-header">' +
                            '<div class="card-title"><span style="font-style: normal;">📈</span> FPS 趋势</div>' +
                        '</div>' +
                        '<div class="chart-container">' +
                            '<div class="chart-bars" id="fpsChart"></div>' +
                        '</div>' +
                    '</div>' +
                    '<div class="card">' +
                        '<div class="card-header">' +
                            '<div class="card-title"><span style="font-style: normal;">💾</span> 内存趋势</div>' +
                        '</div>' +
                        '<div class="chart-container">' +
                            '<div class="chart-bars" id="memoryChart"></div>' +
                        '</div>' +
                    '</div>' +
                '</div>' +
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">🎞️</span> 最近帧</div>' +
                        '<button class="btn btn-sm" onclick="loadView(\'frames\')">查看全部</button>' +
                    '</div>' +
                    '<div class="frame-list" id="dashFrames">' +
                        '<div class="empty-state">' +
                            '<div class="icon">🎞️</div>' +
                            '<p>暂无帧数据</p>' +
                        '</div>' +
                    '</div>' +
                '</div>';
            updateMetricsDisplay();
            updateFramesDisplay();
            updateCharts();
        }

        function updateMetricsDisplay() {
            const m = AppState.metrics;
            if (!m) return;

            const updateElement = (id, value) => {
                const el = document.getElementById(id);
                if (el) el.textContent = value;
            };

            updateElement('dashFps', m.fps?.toFixed(1) || '--');
            updateElement('dashFrameTime', m.frameTime || '--');
            updateElement('dashMemory', formatBytes(m.memoryUsage));
            updateElement('dashComponents', m.componentCount || '--');
        }

        function updateFramesDisplay() {
            const container = document.getElementById('dashFrames');
            if (!container) return;

            const frames = AppState.frames.slice(0, 5);
            if (frames.length === 0) {
                container.innerHTML = '<div class="empty-state"><div class="icon">🎞️</div><p>暂无帧数据</p></div>';
                return;
            }

            container.innerHTML = frames.map(f =>
                '<div class="frame-item">' +
                    '<span class="frame-id">#' + f.frameId + '</span>' +
                    '<div class="frame-info">' +
                        '<span>📋 ' + (f.eventCount || 0) + ' 事件</span>' +
                        '<span>🔄 ' + (f.mutationCount || 0) + ' 变更</span>' +
                        '<span>📐 ' + (f.layoutCount || 0) + ' 布局</span>' +
                    '</div>' +
                    '<span class="frame-timestamp">' + formatTimestamp(f.timestamp) + '</span>' +
                '</div>'
            ).join('');
        }

        function updateCharts() {
            // FPS Chart
            const fpsContainer = document.getElementById('fpsChart');
            if (fpsContainer) {
                const maxFps = Math.max(...AppState.fpsHistory, 60);
                fpsContainer.innerHTML = AppState.fpsHistory.map((fps, i) => {
                    const height = Math.max(4, (fps / maxFps) * 100);
                    const color = fps >= 50 ? 'var(--accent-green)' : fps >= 30 ? 'var(--accent-yellow)' : 'var(--accent-red)';
                    return '<div class="chart-bar" style="height: ' + height + '%; background: ' + color + ';" title="' + fps.toFixed(1) + ' FPS"></div>';
                }).join('');
            }

            // Memory Chart
            const memContainer = document.getElementById('memoryChart');
            if (memContainer) {
                const maxMem = Math.max(...AppState.memoryHistory, 1);
                memContainer.innerHTML = AppState.memoryHistory.map((mem, i) => {
                    const height = Math.max(4, (mem / maxMem) * 100);
                    return '<div class="chart-bar" style="height: ' + height + '%;" title="' + formatBytes(mem) + '"></div>';
                }).join('');
            }
        }

        function renderMetrics(metrics) {
            const content = document.getElementById('mainContent');
            if (!metrics) {
                content.innerHTML = '<div class="empty-state"><div class="icon">📈</div><p>暂无性能指标</p></div>';
                return;
            }

            content.innerHTML =
                '<div class="metrics-grid">' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">⚡</span> FPS</div>' +
                        '<div class="metric-card-value fps">' + (metrics.fps?.toFixed(1) || '--') + '</div>' +
                        '<div class="metric-card-delta">帧每秒</div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">⏱️</span> 帧时间</div>' +
                        '<div class="metric-card-value time">' + (metrics.frameTime || '--') + '</div>' +
                        '<div class="metric-card-delta">毫秒/帧</div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">📐</span> 布局时间</div>' +
                        '<div class="metric-card-value time">' + (metrics.layoutTime || '--') + '</div>' +
                        '<div class="metric-card-delta">毫秒</div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">🎨</span> 绘制时间</div>' +
                        '<div class="metric-card-value time">' + (metrics.paintTime || '--') + '</div>' +
                        '<div class="metric-card-delta">毫秒</div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">💾</span> 内存</div>' +
                        '<div class="metric-card-value memory">' + formatBytes(metrics.memoryUsage) + '</div>' +
                        '<div class="metric-card-delta">' + (metrics.memoryUsage / 1024 / 1024).toFixed(2) + ' MB</div>' +
                    '</div>' +
                    '<div class="metric-card">' +
                        '<div class="metric-card-label"><span style="font-style: normal;">🧩</span> 组件</div>' +
                        '<div class="metric-card-value count">' + (metrics.componentCount || '--') + '</div>' +
                        '<div class="metric-card-delta">活动组件</div>' +
                    '</div>' +
                '</div>' +
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">📊</span> 性能详情</div>' +
                    '</div>' +
                    '<div class="table-container">' +
                        '<table>' +
                            '<thead>' +
                                '<tr>' +
                                    '<th>指标</th>' +
                                    '<th>值</th>' +
                                    '<th>说明</th>' +
                                '</tr>' +
                            '</thead>' +
                            '<tbody>' +
                                '<tr>' +
                                    '<td>目标帧率</td>' +
                                    '<td>60 FPS</td>' +
                                    '<td>目标帧时间: 16.67ms</td>' +
                                '</tr>' +
                                '<tr>' +
                                    '<td>当前帧率</td>' +
                                    '<td class="' + (metrics.fps >= 50 ? 'text-green' : metrics.fps >= 30 ? 'text-yellow' : 'text-red') + '">' + (metrics.fps?.toFixed(1) || '--') + ' FPS</td>' +
                                    '<td>' + (metrics.fps >= 50 ? '✅ 良好' : metrics.fps >= 30 ? '⚠️ 一般' : '❌ 需要优化') + '</td>' +
                                '</tr>' +
                                '<tr>' +
                                    '<td>帧时间</td>' +
                                    '<td>' + (metrics.frameTime || 0) + ' ms</td>' +
                                    '<td>每帧渲染时间</td>' +
                                '</tr>' +
                                '<tr>' +
                                    '<td>总帧数</td>' +
                                    '<td>' + (metrics.frameCount || 0) + '</td>' +
                                    '<td>已捕获的帧总数</td>' +
                                '</tr>' +
                            '</tbody>' +
                        '</table>' +
                    '</div>' +
                '</div>' +
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">📈</span> FPS 趋势图</div>' +
                    '</div>' +
                    '<div class="chart-container">' +
                        '<div class="chart-bars" id="fpsChartDetail"></div>' +
                    '</div>' +
                '</div>';

            // Render FPS chart
            const chart = document.getElementById('fpsChartDetail');
            if (chart) {
                const maxFps = Math.max(...AppState.fpsHistory, 60);
                chart.innerHTML = AppState.fpsHistory.map((fps, i) => {
                    const height = Math.max(4, (fps / maxFps) * 100);
                    const color = fps >= 50 ? 'var(--accent-green)' : fps >= 30 ? 'var(--accent-yellow)' : 'var(--accent-red)';
                    return '<div class="chart-bar" style="height: ' + height + '%; background: ' + color + ';" title="Frame ' + i + ': ' + fps.toFixed(1) + ' FPS"></div>';
                }).join('');
            }
        }

        function renderFrames(frames) {
            const content = document.getElementById('mainContent');
            if (!frames || frames.length === 0) {
                content.innerHTML = '<div class="empty-state"><div class="icon">🎞️</div><p>暂无帧数据</p></div>';
                return;
            }

            AppState.frames = frames;

            content.innerHTML =
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">🎞️</span> 帧列表</div>' +
                        '<div class="card-actions">' +
                            '<span class="badge badge-blue">共 ' + frames.length + ' 帧</span>' +
                        '</div>' +
                    '</div>' +
                    '<div class="frame-list">' +
                        frames.map(f =>
                            '<div class="frame-item" onclick="showFrameDetail(' + f.frameId + ')">' +
                                '<span class="frame-id">#' + f.frameId + '</span>' +
                                '<div class="frame-info">' +
                                    '<span>📋 ' + (f.eventCount || 0) + ' 事件</span>' +
                                    '<span>🔄 ' + (f.mutationCount || 0) + ' 变更</span>' +
                                    '<span>📐 ' + (f.layoutCount || 0) + ' 布局</span>' +
                                    '<span>🎨 ' + (f.repaintCount || 0) + ' 重绘</span>' +
                                '</div>' +
                                '<span class="frame-timestamp">' + formatTimestamp(f.timestamp) + '</span>' +
                            '</div>'
                        ).join('') +
                    '</div>' +
                '</div>';
        }

        function renderComponents(components) {
            const content = document.getElementById('mainContent');
            if (!components || components.length === 0) {
                content.innerHTML = '<div class="empty-state"><div class="icon">🧩</div><p>暂无组件</p></div>';
                return;
            }

            AppState.components = components;

            content.innerHTML =
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">🧩</span> 组件树</div>' +
                        '<div class="card-actions">' +
                            '<span class="badge badge-blue">共 ' + components.length + ' 个组件</span>' +
                        '</div>' +
                    '</div>' +
                    '<div class="search-box">' +
                        '<span class="icon">🔍</span>' +
                        '<input type="text" placeholder="搜索组件..." oninput="filterComponents(this.value)">' +
                    '</div>' +
                    '<div class="component-tree" id="componentTree">' +
                        renderComponentTree(components) +
                    '</div>' +
                '</div>';
        }

        function renderComponentTree(components) {
            const comps = Array.isArray(components) ? components : Object.values(components);
            return comps.map(c =>
                '<div class="tree-node" onclick="showComponentDetail(\'' + (c.id || '') + '\')">' +
                    '<span class="component-type">' + (c.type || 'Unknown') + '</span>' +
                    '<span class="component-id">' + (c.id || '') + '</span>' +
                    (c.visible ? '<span class="badge badge-green">可见</span>' : '<span class="badge badge-red">隐藏</span>') +
                '</div>'
            ).join('');
        }

        function renderSnapshots(snapshots) {
            const content = document.getElementById('mainContent');
            if (!snapshots || snapshots.length === 0) {
                content.innerHTML = '<div class="empty-state"><div class="icon">📸</div><p>暂无快照</p></div>';
                return;
            }

            AppState.snapshots = snapshots;

            content.innerHTML =
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">📸</span> 快照列表</div>' +
                        '<div class="card-actions">' +
                            '<span class="badge badge-blue">共 ' + snapshots.length + ' 个快照</span>' +
                            '<button class="btn btn-sm" onclick="loadView(\'diff\')">对比快照</button>' +
                        '</div>' +
                    '</div>' +
                    '<div class="snapshot-grid">' +
                        snapshots.map(s =>
                            '<div class="snapshot-card" onclick="showSnapshotDetail(' + s.frame_id + ')">' +
                                '<div class="snapshot-card-header">' +
                                    '<span class="snapshot-frame-id">#' + s.frame_id + '</span>' +
                                    '<span class="snapshot-badge">' + s.components + ' 组件</span>' +
                                '</div>' +
                                '<div class="snapshot-info">' +
                                    '<div>🆔 ' + (s.id ? s.id.substring(0, 16) : '') + '...</div>' +
                                    '<div>⏰ ' + formatTimestamp(s.timestamp) + '</div>' +
                                '</div>' +
                            '</div>'
                        ).join('') +
                    '</div>' +
                '</div>';
        }

        function renderDiff() {
            const content = document.getElementById('mainContent');

            // Get available frames from snapshots
            const frames = AppState.snapshots.map(s => s.frame_id);
            const fromFrame = frames[0] || 0;
            const toFrame = frames[frames.length - 1] || 0;

            content.innerHTML =
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">🔍</span> 快照差异对比</div>' +
                    '</div>' +
                    '<div style="display: flex; gap: 12px; align-items: center; margin-bottom: 16px;">' +
                        '<div style="flex: 1;">' +
                            '<label style="font-size: 12px; color: var(--text-secondary); display: block; margin-bottom: 4px;">起始帧</label>' +
                            '<input type="number" id="fromFrame" value="' + fromFrame + '" style="width: 100%; padding: 8px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 4px; color: var(--text-primary);">' +
                        '</div>' +
                        '<div style="padding-top: 20px;">→</div>' +
                        '<div style="flex: 1;">' +
                            '<label style="font-size: 12px; color: var(--text-secondary); display: block; margin-bottom: 4px;">结束帧</label>' +
                            '<input type="number" id="toFrame" value="' + toFrame + '" style="width: 100%; padding: 8px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 4px; color: var(--text-primary);">' +
                        '</div>' +
                        '<div style="padding-top: 20px;">' +
                            '<button class="btn btn-primary" onclick="loadDiff()">对比</button>' +
                        '</div>' +
                    '</div>' +
                    '<div id="diffResult" class="diff-container">' +
                        '<div class="empty-state">' +
                            '<div class="icon">🔍</div>' +
                            '<p>选择帧范围进行对比</p>' +
                        '</div>' +
                    '</div>' +
                '</div>';

            // Load snapshots for frame selection
            fetch('/api/snapshots')
                .then(r => r.json())
                .then(snapshots => {
                    AppState.snapshots = snapshots;
                });
        }

        function loadDiff() {
            const from = document.getElementById('fromFrame').value;
            const to = document.getElementById('toFrame').value;

            if (!from || !to) {
                alert('请输入帧ID');
                return;
            }

            fetch('/api/diff?from=' + from + '&to=' + to)
                .then(r => r.json())
                .then(data => {
                    displayDiff(data);
                })
                .catch(err => {
                    document.getElementById('diffResult').innerHTML =
                        '<div class="empty-state"><div class="icon">⚠️</div><p>加载失败: ' + err.message + '</p></div>';
                });
        }

        function displayDiff(data) {
            const container = document.getElementById('diffResult');
            if (!data.changes || data.changes.length === 0) {
                container.innerHTML = '<div class="empty-state"><div class="icon">✅</div><p>没有差异</p></div>';
                return;
            }

            container.innerHTML =
                '<div style="margin-bottom: 12px;">' +
                    '<span class="badge badge-blue">帧 #' + data.from + '</span>' +
                    '<span style="color: var(--text-secondary);"> → </span>' +
                    '<span class="badge badge-blue">帧 #' + data.to + '</span>' +
                    '<span class="badge badge-yellow">' + data.changes.length + ' 处变更</span>' +
                '</div>' +
                data.changes.map(c =>
                    '<div class="diff-item ' + c.type + '">' +
                        '<span class="diff-icon">' + (c.type === 'added' ? '+' : c.type === 'removed' ? '-' : '~') + '</span>' +
                        '<span>' + (c.node_id || c.NodeID || 'unknown') + '</span>' +
                        (c.path ? '<span style="color: var(--text-secondary); margin-left: 8px;">' + c.path + '</span>' : '') +
                    '</div>'
                ).join('');
        }

        function renderReport(report) {
            const content = document.getElementById('mainContent');
            if (!report) {
                content.innerHTML = '<div class="empty-state"><div class="icon">📋</div><p>暂无报告</p></div>';
                return;
            }

            content.innerHTML =
                '<div class="card">' +
                    '<div class="card-header">' +
                        '<div class="card-title"><span style="font-style: normal;">📋</span> 调试报告</div>' +
                        '<div class="card-actions">' +
                            '<button class="btn btn-sm" onclick="exportReport()">导出</button>' +
                        '</div>' +
                    '</div>' +
                    '<div class="detail-panel">' +
                        '<div class="detail-row">' +
                            '<span class="detail-label">生成时间</span>' +
                            '<span class="detail-value">' + formatTimestamp(report.generatedAt) + '</span>' +
                        '</div>' +
                        '<div class="detail-row">' +
                            '<span class="detail-label">总帧数</span>' +
                            '<span class="detail-value">' + (report.frames?.length || 0) + '</span>' +
                        '</div>' +
                        '<div class="detail-row">' +
                            '<span class="detail-label">总组件数</span>' +
                            '<span class="detail-value">' + (report.components?.length || 0) + '</span>' +
                        '</div>' +
                        '<div class="detail-row">' +
                            '<span class="detail-label">FPS</span>' +
                            '<span class="detail-value">' + (report.metrics?.fps?.toFixed(1) || '--') + '</span>' +
                        '</div>' +
                        '<div class="detail-row">' +
                            '<span class="detail-label">内存使用</span>' +
                            '<span class="detail-value">' + formatBytes(report.metrics?.memoryUsage) + '</span>' +
                        '</div>' +
                    '</div>' +
                '</div>' +
                (report.metrics ?
                    '<div class="card">' +
                        '<div class="card-header">' +
                            '<div class="card-title"><span style="font-style: normal;">📊</span> 性能摘要</div>' +
                        '</div>' +
                        '<div class="metrics-grid">' +
                            '<div class="metric-card">' +
                                '<div class="metric-card-label">平均 FPS</div>' +
                                '<div class="metric-card-value fps">' + (report.metrics.fps?.toFixed(1) || '--') + '</div>' +
                            '</div>' +
                            '<div class="metric-card">' +
                                '<div class="metric-card-label">帧时间</div>' +
                                '<div class="metric-card-value time">' + (report.metrics.frameTime || '--') + ' ms</div>' +
                            '</div>' +
                            '<div class="metric-card">' +
                                '<div class="metric-card-label">内存峰值</div>' +
                                '<div class="metric-card-value memory">' + formatBytes(report.metrics.memoryUsage) + '</div>' +
                            '</div>' +
                            '<div class="metric-card">' +
                                '<div class="metric-card-label">组件数量</div>' +
                                '<div class="metric-card-value count">' + (report.metrics.componentCount || '--') + '</div>' +
                            '</div>' +
                        '</div>' +
                    '</div>'
                : '');
        }

        // Utility Functions
        function formatBytes(bytes) {
            if (!bytes || bytes === 0) return '0 B';
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
            return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
        }

        function formatTimestamp(timestamp) {
            if (!timestamp) return '--';
            const date = new Date(timestamp);
            return date.toLocaleTimeString('zh-CN');
        }

        function showFrameDetail(frameId) {
            console.log('Show frame detail:', frameId);
            // TODO: Implement frame detail view
        }

        function showComponentDetail(componentId) {
            console.log('Show component detail:', componentId);
            // TODO: Implement component detail view
        }

        function showSnapshotDetail(frameId) {
            console.log('Show snapshot detail:', frameId);
            // TODO: Implement snapshot detail view
        }

        function filterComponents(query) {
            const tree = document.getElementById('componentTree');
            if (!tree) return;

            const nodes = tree.querySelectorAll('.tree-node');
            query = query.toLowerCase();

            nodes.forEach(node => {
                const text = node.textContent.toLowerCase();
                node.style.display = text.includes(query) ? '' : 'none';
            });
        }

        function exportReport() {
            fetch('/api/export')
                .then(r => r.json())
                .then(data => {
                    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = 'mint-devtools-export-' + Date.now() + '.json';
                    a.click();
                    URL.revokeObjectURL(url);
                });
        }

        // Initialize
        document.addEventListener('DOMContentLoaded', () => {
            // Setup sidebar navigation
            document.querySelectorAll('.sidebar-item').forEach(item => {
                item.addEventListener('click', () => loadView(item.dataset.view));
            });

            // Connect WebSocket
            connect();

            // Load initial view
            loadView('dashboard');
        });
    </script>
</body>
</html>`
}
