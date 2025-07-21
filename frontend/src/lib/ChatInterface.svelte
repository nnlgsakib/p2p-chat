<script>
	import { onMount, createEventDispatcher } from 'svelte';
	import { fly } from 'svelte/transition';

	export let connectedPeers = [];
	export let ws = null;
	export let nodeInfo = null;
	export let messagesByPeer = {};

	const dispatch = createEventDispatcher();

	let selectedPeer = null;
	let newMessage = '';
	let loading = false;
	let messages = [];

	$: {
		if (selectedPeer && messagesByPeer[selectedPeer]) {
			messages = messagesByPeer[selectedPeer];
		} else {
			messages = [];
		}
	}

	function selectPeer(peer) {
		selectedPeer = peer;
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify({
				type: 'get_chat_history',
				peer_id: peer
			}));
		}
	}

	async function sendMessage() {
		if (!newMessage.trim() || !selectedPeer || loading || !ws) return;

		loading = true;
		try {
			//This is not ideal, we should be using the websocket to send the message
			//but the backend is not setup to handle it.
			//ws.send(JSON.stringify({
			//	type: 'send_private_message',
			//	peer_id: selectedPeer,
			//	message: newMessage.trim()
			//}));
			
			const response = await fetch('/chat/private/send', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					peer_id: selectedPeer,
					message: newMessage.trim()
				})
			});

			if (response.ok) {
				const sentMessage = {
					id: new Date().getTime().toString(), // Temporary ID
					content: newMessage.trim(),
					timestamp: new Date().toISOString(),
					sent: true,
					sender: nodeInfo.peer_id,
				};
				messagesByPeer[selectedPeer] = [...(messagesByPeer[selectedPeer] || []), sentMessage];
				newMessage = '';
			} else {
				console.error('Failed to send message');
			}

		} catch (error) {
			console.error('Error sending message:', error);
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(event) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			sendMessage();
		}
	}
</script>

<div class="card">
	<h2>Private Chat</h2>
	
	<div class="chat-container">
		<div class="chat-sidebar">
			<h3>Connected Peers</h3>
			{#if connectedPeers.length === 0}
				<p>No connected peers</p>
			{:else}
				<ul class="peer-list">
					{#each connectedPeers as peer}
						<li 
							class="peer-item {selectedPeer === peer ? 'active' : ''}"
							on:click={() => selectPeer(peer)}
						>
							<span class="status-indicator status-online"></span>
							{peer.slice(0, 12)}...
						</li>
					{/each}
				</ul>
			{/if}
		</div>

		<div class="chat-main">
			{#if selectedPeer}
				<div class="chat-header">
					<h4>Chat with {selectedPeer.slice(0, 12)}...</h4>
				</div>

				<div class="chat-messages">
					{#if messages.length === 0}
						<p class="no-messages">No messages yet. Start the conversation!</p>
					{:else}
						{#each messages as message (message.id)}
							<div class="message {message.sent ? 'sent' : 'received'}" in:fly={{ y: 20, duration: 300 }}>
								<div class="message-sender">{message.sent ? 'You' : message.sender.slice(0,12) + '...'}</div>
								<div class="message-content">{message.content}</div>
								<div class="message-timestamp">{new Date(message.timestamp).toLocaleTimeString()}</div>
							</div>
						{/each}
					{/if}
				</div>

				<div class="chat-input">
					<div class="input-group">
						<textarea
							bind:value={newMessage}
							on:keypress={handleKeyPress}
							placeholder="Type your message..."
							rows="2"
							disabled={loading || !selectedPeer}
						></textarea>
						<button 
							class="btn"
							on:click={sendMessage}
							disabled={!newMessage.trim() || loading || !selectedPeer}
						>
							{loading ? 'Sending...' : 'Send'}
						</button>
					</div>
				</div>
			{:else}
				<div class="no-selection">
					<p>Select a peer to start chatting</p>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.chat-header {
		padding: 15px 20px;
		border-bottom: 1px solid #ddd;
		background-color: #f8f9fa;
	}

	.chat-header h4 {
		margin: 0;
		font-size: 16px;
	}

	.no-messages {
		text-align: center;
		color: #6c757d;
		font-style: italic;
		margin-top: 50px;
	}

	.no-selection {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		color: #6c757d;
		font-style: italic;
	}

	.input-group {
		display: flex;
		gap: 10px;
		align-items: flex-end;
	}

	.input-group textarea {
		flex: 1;
		padding: 10px;
		border: 1px solid #ddd;
		border-radius: 4px;
		resize: vertical;
		min-height: 40px;
		font-family: inherit;
	}

	.input-group textarea:focus {
		outline: none;
		border-color: #007bff;
		box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
	}

	h2, h3 {
		margin-top: 0;
	}

	.message-timestamp {
		font-size: 10px;
		opacity: 0.7;
		margin-top: 5px;
		text-align: right;
	}
</style>

