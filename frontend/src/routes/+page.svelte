<script>
	import { onMount } from 'svelte';
	import ChatInterface from '../lib/ChatInterface.svelte';
	import PeerManager from '../lib/PeerManager.svelte';
	import GroupManager from '../lib/GroupManager.svelte';
	import FileManager from '../lib/FileManager.svelte';
	import NodeInfo from '../lib/NodeInfo.svelte';

	let activeTab = 'chat';
	let ws = null;
	let connected = false;
	let nodeInfo = null;
	let connectedPeers = [];
	let groups = [];
	let receivedFiles = [];
	let chatMessages = {}; // Store messages per peer

	const tabs = [
		{ id: 'chat', label: 'Chat' },
		{ id: 'peers', label: 'Peers' },
		{ id: 'groups', label: 'Groups' },
		{ id: 'files', label: 'Files' },
		{ id: 'info', label: 'Node Info' }
	];

	onMount(async () => {
		try {
			const response = await fetch('/api/ports');
			const ports = await response.json();
			connectWebSocket(ports.ws_port);
		} catch (e) {
			console.error("Could not fetch ports, using default", e);
			connectWebSocket(8081);
		}
	});

	function connectWebSocket(port) {
		try {
			const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			const wsHost = window.location.hostname;
			ws = new WebSocket(`${wsProtocol}//${wsHost}:${port}/ws`);
			
			ws.onopen = () => {
				connected = true;
				console.log('WebSocket connected');
				requestNodeInfo();
				requestConnectedPeers();
				requestGroups();
				requestReceivedFiles();
			};

			ws.onmessage = (event) => {
				const data = JSON.parse(event.data);
				handleWebSocketMessage(data);
			};

			ws.onclose = () => {
				connected = false;
				console.log('WebSocket disconnected');
				// Attempt to reconnect after 3 seconds
				setTimeout(() => connectWebSocket(port), 3000);
			};

			ws.onerror = (error) => {
				console.error('WebSocket error:', error);
			};
		} catch (error) {
			console.error('Failed to connect WebSocket:', error);
			setTimeout(() => connectWebSocket(port), 3000);
		}
	}

	function handleWebSocketMessage(data) {
		switch (data.type) {
			case 'peer_info':
				nodeInfo = data;
				break;
			case 'connected_peers':
				connectedPeers = data.peers;
				break;
			case 'groups':
				groups = data.groups;
				break;
			case 'received_files':
				receivedFiles = data.files;
				break;
			case 'new_message':
				if (data.message_type === 'private') {
					const peerId = data.sender_id === nodeInfo.peer_id ? data.recipient_id : data.sender_id;
					if (!chatMessages[peerId]) {
						chatMessages[peerId] = [];
					}
					chatMessages[peerId] = [...chatMessages[peerId], {
						id: data.timestamp,
						sender: data.sender_id,
						content: data.message,
						timestamp: new Date(data.timestamp * 1000),
						sent: data.sender_id === nodeInfo.peer_id
					}];
				}
				break;
			case 'chat_history':
				chatMessages[data.peer_id] = data.history.map(msg => ({
					id: msg.timestamp,
					sender: msg.sender_id,
					content: msg.content,
					timestamp: new Date(msg.timestamp * 1000),
					sent: msg.sender_id === nodeInfo.peer_id
				}));
				break;
			case 'error':
				console.error('WebSocket error:', data.error);
				break;
		}
	}

	function requestNodeInfo() {
		if (ws && connected) {
			ws.send(JSON.stringify({ type: 'get_peer_info' }));
		}
	}

	function requestConnectedPeers() {
		if (ws && connected) {
			ws.send(JSON.stringify({ type: 'get_connected_peers' }));
		}
	}

	function requestGroups() {
		if (ws && connected) {
			ws.send(JSON.stringify({ type: 'get_groups' }));
		}
	}

	function requestReceivedFiles() {
		if (ws && connected) {
			ws.send(JSON.stringify({ type: 'get_received_files' }));
		}
	}

	function refreshData() {
		requestNodeInfo();
		requestConnectedPeers();
		requestGroups();
		requestReceivedFiles();
	}
</script>

<div class="container">
	<header>
		<h1>P2P Chat Application</h1>
		<div class="status">
			<span class="status-indicator {connected ? 'status-online' : 'status-offline'}"></span>
			{connected ? 'Connected' : 'Disconnected'}
			<button class="btn btn-secondary" on:click={refreshData} disabled={!connected}>
				Refresh
			</button>
		</div>
	</header>

	<div class="tabs">
		{#each tabs as tab}
			<div 
				class="tab {activeTab === tab.id ? 'active' : ''}"
				on:click={() => activeTab = tab.id}
			>
				{tab.label}
			</div>
		{/each}
	</div>

	<div class="tab-content">
		{#if activeTab === 'chat'}
			<ChatInterface {connectedPeers} {ws} {nodeInfo} bind:messagesByPeer={chatMessages} />
		{:else if activeTab === 'peers'}
			<PeerManager {connectedPeers} on:refresh={refreshData} />
		{:else if activeTab === 'groups'}
			<GroupManager {groups} {nodeInfo} on:refresh={refreshData} />
		{:else if activeTab === 'files'}
			<FileManager {receivedFiles} {connectedPeers} on:refresh={refreshData} />
		{:else if activeTab === 'info'}
			<NodeInfo {nodeInfo} {connectedPeers} />
		{/if}
	</div>
</div>

<style>
	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 30px;
		padding-bottom: 20px;
		border-bottom: 1px solid #ddd;
	}

	h1 {
		margin: 0;
		color: #333;
	}

	.status {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 14px;
	}

	.tab-content {
		min-height: 600px;
	}
</style>

