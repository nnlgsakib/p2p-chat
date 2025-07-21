<script>
	import { createEventDispatcher } from 'svelte';
	
	export let receivedFiles = [];
	export let connectedPeers = [];
	
	const dispatch = createEventDispatcher();
	
	let showSendModal = false;
	let selectedPeer = '';
	let selectedFile = null;
	let loading = false;

	function handleFileSelect(event) {
		const file = event.target.files[0];
		if (file) {
			selectedFile = file;
		}
	}

	async function sendFile() {
		if (!selectedFile || !selectedPeer || loading) return;

		loading = true;
		try {
			// In a real implementation, you would upload the file first
			// and then send the file path to the API
			const response = await fetch('/file/send', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					peer_id: selectedPeer,
					file_path: selectedFile.name // This would be the actual file path
				})
			});

			if (response.ok) {
				showSendModal = false;
				selectedFile = null;
				selectedPeer = '';
				dispatch('refresh');
			} else {
				console.error('Failed to send file');
			}
		} catch (error) {
			console.error('Error sending file:', error);
		} finally {
			loading = false;
		}
	}

	function closeModal() {
		showSendModal = false;
		selectedFile = null;
		selectedPeer = '';
	}

	function formatFileSize(bytes) {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}
</script>

<div class="card">
	<div class="header">
		<h2>File Management</h2>
		<button class="btn" on:click={() => showSendModal = true}>
			Send File
		</button>
	</div>

	<div class="files-section">
		<h3>Received Files ({receivedFiles.length})</h3>
		{#if receivedFiles.length === 0}
			<p class="no-files">No files received yet.</p>
		{:else}
			<div class="files-grid">
				{#each receivedFiles as file}
					<div class="file-card">
						<div class="file-icon">
							📄
						</div>
						<div class="file-info">
							<div class="file-name">{file}</div>
							<div class="file-details">
								<span class="file-size">Size: Unknown</span>
								<span class="file-date">Received: Recently</span>
							</div>
						</div>
						<div class="file-actions">
							<button class="btn btn-secondary">
								Download
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Send File Modal -->
{#if showSendModal}
	<div class="modal" on:click={closeModal}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3 class="modal-title">Send File</h3>
			</div>
			
			<div class="form-group">
				<label class="form-label" for="recipient">
					Recipient
				</label>
				<select
					id="recipient"
					class="form-input"
					bind:value={selectedPeer}
					disabled={loading}
				>
					<option value="">Select a peer...</option>
					{#each connectedPeers as peer}
						<option value={peer}>{peer.slice(0, 12)}...</option>
					{/each}
				</select>
			</div>

			<div class="form-group">
				<label class="form-label" for="file-input">
					File
				</label>
				<input
					id="file-input"
					type="file"
					class="form-input"
					on:change={handleFileSelect}
					disabled={loading}
				/>
				{#if selectedFile}
					<div class="selected-file">
						<strong>Selected:</strong> {selectedFile.name} ({formatFileSize(selectedFile.size)})
					</div>
				{/if}
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeModal} disabled={loading}>
					Cancel
				</button>
				<button 
					class="btn" 
					on:click={sendFile}
					disabled={!selectedFile || !selectedPeer || loading}
				>
					{loading ? 'Sending...' : 'Send File'}
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

	.files-section h3 {
		margin-bottom: 15px;
		color: #333;
	}

	.no-files {
		text-align: center;
		color: #6c757d;
		font-style: italic;
		padding: 40px 20px;
	}

	.files-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 15px;
	}

	.file-card {
		background: #f8f9fa;
		border: 1px solid #ddd;
		border-radius: 8px;
		padding: 15px;
		display: flex;
		align-items: center;
		gap: 15px;
	}

	.file-icon {
		font-size: 24px;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #e9ecef;
		border-radius: 6px;
	}

	.file-info {
		flex: 1;
	}

	.file-name {
		font-weight: 500;
		margin-bottom: 5px;
		word-break: break-word;
	}

	.file-details {
		font-size: 12px;
		color: #6c757d;
	}

	.file-details span {
		display: block;
		margin-bottom: 2px;
	}

	.file-actions {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}

	.selected-file {
		margin-top: 10px;
		padding: 10px;
		background: #e9ecef;
		border-radius: 4px;
		font-size: 14px;
	}

	h2 {
		margin: 0;
	}

	select.form-input {
		cursor: pointer;
	}
</style>

