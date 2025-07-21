<script>
	import { createEventDispatcher } from 'svelte';
	
	export let connectedPeers = [];
	
	const dispatch = createEventDispatcher();
	
	let showConnectModal = false;
	let showSearchModal = false;
	let connectAddress = '';
	let searchQuery = '';
	let searchResults = [];
	let loading = false;

	async function connectToPeer() {
		if (!connectAddress.trim() || loading) return;

		loading = true;
		try {
			const response = await fetch('/peer/connect', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					peer_multiaddr: connectAddress.trim()
				})
			});

			if (response.ok) {
				showConnectModal = false;
				connectAddress = '';
				dispatch('refresh');
			} else {
				console.error('Failed to connect to peer');
			}
		} catch (error) {
			console.error('Error connecting to peer:', error);
		} finally {
			loading = false;
		}
	}

	async function searchPeer() {
		if (!searchQuery.trim() || loading) return;

		loading = true;
		try {
			const response = await fetch(`/peer/search?query=${encodeURIComponent(searchQuery.trim())}`);
			
			if (response.ok) {
				const data = await response.json();
				if (data.peer_id && data.addrs && data.addrs.length > 0) {
					// Construct the multiaddress
					const fullAddress = `${data.addrs[0]}/p2p/${data.peer_id}`;
					searchResults = [fullAddress];
				} else {
					searchResults = [];
				}
			} else {
				console.error('Failed to search peer');
				searchResults = [];
			}
		} catch (error) {
			console.error('Error searching peer:', error);
			searchResults = [];
		} finally {
			loading = false;
		}
	}

	function closeModals() {
		showConnectModal = false;
		showSearchModal = false;
		connectAddress = '';
		searchQuery = '';
		searchResults = [];
	}
</script>

<div class="card">
	<div class="header">
		<h2>Peer Management</h2>
		<div class="actions">
			<button class="btn" on:click={() => showConnectModal = true}>
				Connect to Peer
			</button>
			<button class="btn btn-secondary" on:click={() => showSearchModal = true}>
				Search Peers
			</button>
		</div>
	</div>

	<div class="peer-section">
		<h3>Connected Peers ({connectedPeers.length})</h3>
		{#if connectedPeers.length === 0}
			<p class="no-peers">No connected peers. Use the buttons above to connect to peers.</p>
		{:else}
			<div class="peer-grid">
				{#each connectedPeers as peer}
					<div class="peer-card">
						<div class="peer-id">
							<strong>Peer ID:</strong> {peer}
						</div>
						<div class="peer-status">
							<span class="status-indicator status-online"></span>
							Online
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Connect Modal -->
{#if showConnectModal}
	<div class="modal" on:click={closeModals}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3 class="modal-title">Connect to Peer</h3>
			</div>
			
			<div class="form-group">
				<label class="form-label" for="connect-address">
					Peer Multiaddress
				</label>
				<input
					id="connect-address"
					type="text"
					class="form-input"
					bind:value={connectAddress}
					placeholder="/ip4/127.0.0.1/tcp/4001/p2p/12D3KooW..."
					disabled={loading}
				/>
				<small>Enter the full multiaddress of the peer you want to connect to</small>
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeModals} disabled={loading}>
					Cancel
				</button>
				<button 
					class="btn" 
					on:click={connectToPeer}
					disabled={!connectAddress.trim() || loading}
				>
					{loading ? 'Connecting...' : 'Connect'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Search Modal -->
{#if showSearchModal}
	<div class="modal" on:click={closeModals}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3 class="modal-title">Search Peers</h3>
			</div>
			
			<div class="form-group">
				<label class="form-label" for="search-query">
					Search Query
				</label>
				<div class="search-input-group">
					<input
						id="search-query"
						type="text"
						class="form-input"
						bind:value={searchQuery}
						placeholder="Enter username or peer ID..."
						disabled={loading}
					/>
					<button 
						class="btn" 
						on:click={searchPeer}
						disabled={!searchQuery.trim() || loading}
					>
						{loading ? 'Searching...' : 'Search'}
					</button>
				</div>
			</div>

			{#if searchResults.length > 0}
				<div class="search-results">
					<h4>Search Results</h4>
					{#each searchResults as result}
						<div class="search-result-item">
							<span>{result}</span>
							<button class="btn btn-secondary" on:click={() => connectAddress = result}>
								Use Address
							</button>
						</div>
					{/each}
				</div>
			{:else if searchQuery && !loading}
				<p class="no-results">No peers found for "{searchQuery}"</p>
			{/if}

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeModals}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
	}

	.actions {
		display: flex;
		gap: 10px;
	}

	.peer-section h3 {
		margin-bottom: 15px;
		color: #333;
	}

	.no-peers {
		text-align: center;
		color: #6c757d;
		font-style: italic;
		padding: 40px 20px;
	}

	.peer-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 15px;
	}

	.peer-card {
		background: #f8f9fa;
		border: 1px solid #ddd;
		border-radius: 6px;
		padding: 15px;
	}

	.peer-id {
		font-size: 12px;
		word-break: break-all;
		margin-bottom: 10px;
	}

	.peer-status {
		display: flex;
		align-items: center;
		font-size: 14px;
		color: #28a745;
	}

	.search-input-group {
		display: flex;
		gap: 10px;
	}

	.search-input-group input {
		flex: 1;
	}

	.search-results {
		margin-top: 20px;
	}

	.search-results h4 {
		margin-bottom: 10px;
	}

	.search-result-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px;
		border: 1px solid #ddd;
		border-radius: 4px;
		margin-bottom: 5px;
		font-size: 12px;
		word-break: break-all;
	}

	.no-results {
		text-align: center;
		color: #6c757d;
		font-style: italic;
		margin-top: 20px;
	}

	h2 {
		margin: 0;
	}

	small {
		color: #6c757d;
		font-size: 12px;
	}
</style>

