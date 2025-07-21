<script>
	export let nodeInfo = null;
	export let connectedPeers = [];

	function copyToClipboard(text) {
		navigator.clipboard.writeText(text).then(() => {
			// You could show a toast notification here
			console.log('Copied to clipboard');
		}).catch(err => {
			console.error('Failed to copy: ', err);
		});
	}
</script>

<div class="card">
	<h2>Node Information</h2>
	
	{#if nodeInfo}
		<div class="info-section">
			<h3>Peer Identity</h3>
			<div class="info-item">
				<label>Peer ID:</label>
				<div class="value-with-copy">
					<span class="value">{nodeInfo.peer_id}</span>
					<button 
						class="btn btn-secondary copy-btn"
						on:click={() => copyToClipboard(nodeInfo.peer_id)}
					>
						Copy
					</button>
				</div>
			</div>
		</div>

		<div class="info-section">
			<h3>Network Addresses</h3>
			{#if nodeInfo.addrs && nodeInfo.addrs.length > 0}
				{#each nodeInfo.addrs as addr, index}
					<div class="info-item">
						<label>Address {index + 1}:</label>
						<div class="value-with-copy">
							<span class="value">{addr}</span>
							<button 
								class="btn btn-secondary copy-btn"
								on:click={() => copyToClipboard(addr)}
							>
								Copy
							</button>
						</div>
					</div>
				{/each}
			{:else}
				<p class="no-data">No network addresses available</p>
			{/if}
		</div>

		<div class="info-section">
			<h3>Full Multiaddresses</h3>
			{#if nodeInfo.addrs && nodeInfo.addrs.length > 0}
				{#each nodeInfo.addrs as addr, index}
					<div class="info-item">
						<label>Multiaddr {index + 1}:</label>
						<div class="value-with-copy">
							<span class="value">{addr}/p2p/{nodeInfo.peer_id}</span>
							<button 
								class="btn btn-secondary copy-btn"
								on:click={() => copyToClipboard(`${addr}/p2p/${nodeInfo.peer_id}`)}
							>
								Copy
							</button>
						</div>
					</div>
				{/each}
			{:else}
				<p class="no-data">No multiaddresses available</p>
			{/if}
		</div>
	{:else}
		<div class="loading-state">
			<p>Loading node information...</p>
		</div>
	{/if}

	<div class="info-section">
		<h3>Connection Statistics</h3>
		<div class="stats-grid">
			<div class="stat-item">
				<div class="stat-value">{connectedPeers.length}</div>
				<div class="stat-label">Connected Peers</div>
			</div>
			<div class="stat-item">
				<div class="stat-value">Online</div>
				<div class="stat-label">Node Status</div>
			</div>
		</div>
	</div>

	<div class="info-section">
		<h3>Usage Instructions</h3>
		<div class="instructions">
			<p>
				<strong>To connect to this node:</strong> 
				Share any of the full multiaddresses above with other peers. 
				They can use the "Connect to Peer" feature to establish a connection.
			</p>
			<p>
				<strong>Peer ID:</strong> 
				Your unique identifier on the network. This never changes and 
				identifies your node across sessions.
			</p>
			<p>
				<strong>Network Addresses:</strong> 
				The network locations where your node can be reached. These may 
				change when you restart the node or change networks.
			</p>
		</div>
	</div>
</div>

<style>
	.info-section {
		margin-bottom: 30px;
	}

	.info-section h3 {
		margin-bottom: 15px;
		color: #333;
		border-bottom: 1px solid #ddd;
		padding-bottom: 5px;
	}

	.info-item {
		margin-bottom: 15px;
	}

	.info-item label {
		display: block;
		font-weight: 500;
		margin-bottom: 5px;
		color: #555;
	}

	.value-with-copy {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.value {
		flex: 1;
		font-family: 'Courier New', monospace;
		font-size: 12px;
		background: #f8f9fa;
		padding: 8px 12px;
		border-radius: 4px;
		border: 1px solid #ddd;
		word-break: break-all;
	}

	.copy-btn {
		flex-shrink: 0;
		padding: 8px 12px;
		font-size: 12px;
	}

	.no-data {
		color: #6c757d;
		font-style: italic;
	}

	.loading-state {
		text-align: center;
		padding: 40px 20px;
		color: #6c757d;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 20px;
	}

	.stat-item {
		text-align: center;
		padding: 20px;
		background: #f8f9fa;
		border-radius: 8px;
		border: 1px solid #ddd;
	}

	.stat-value {
		font-size: 24px;
		font-weight: bold;
		color: #007bff;
		margin-bottom: 5px;
	}

	.stat-label {
		font-size: 14px;
		color: #6c757d;
	}

	.instructions {
		background: #e9ecef;
		padding: 20px;
		border-radius: 6px;
		font-size: 14px;
		line-height: 1.5;
	}

	.instructions p {
		margin-bottom: 15px;
	}

	.instructions p:last-child {
		margin-bottom: 0;
	}

	h2 {
		margin-top: 0;
	}
</style>

