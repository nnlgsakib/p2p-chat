<script>
	import { createEventDispatcher } from 'svelte';
	
	export let groups = [];
	export let nodeInfo = null;
	
	const dispatch = createEventDispatcher();
	
	let showCreateModal = false;
	let showAddMemberModal = false;
	let selectedGroup = null;
	let newGroupId = '';
	let newGroupName = '';
	let newMemberId = '';
	let loading = false;

	async function createGroup() {
		if (!newGroupId.trim() || !newGroupName.trim() || !nodeInfo || loading) return;

		loading = true;
		try {
			const response = await fetch('/group/create', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					group_id: newGroupId.trim(),
					group_name: newGroupName.trim(),
					admin_id: nodeInfo.peer_id
				})
			});

			if (response.ok) {
				showCreateModal = false;
				newGroupId = '';
				newGroupName = '';
				dispatch('refresh');
			} else {
				console.error('Failed to create group');
			}
		} catch (error) {
			console.error('Error creating group:', error);
		} finally {
			loading = false;
		}
	}

	async function addMember() {
		if (!newMemberId.trim() || !selectedGroup || loading) return;

		loading = true;
		try {
			const response = await fetch('/group/add_member', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					group_id: selectedGroup.id,
					member_id: newMemberId.trim()
				})
			});

			if (response.ok) {
				showAddMemberModal = false;
				selectedGroup = null;
				newMemberId = '';
				dispatch('refresh');
			} else {
				console.error('Failed to add member');
			}
		} catch (error) {
			console.error('Error adding member:', error);
		} finally {
			loading = false;
		}
	}

	function openAddMemberModal(group) {
		selectedGroup = group;
		showAddMemberModal = true;
	}

	function closeModals() {
		showCreateModal = false;
		showAddMemberModal = false;
		selectedGroup = null;
		newGroupId = '';
		newGroupName = '';
		newMemberId = '';
	}
</script>

<div class="card">
	<div class="header">
		<h2>Group Management</h2>
		<button class="btn" on:click={() => showCreateModal = true}>
			Create Group
		</button>
	</div>

	<div class="groups-section">
		<h3>Your Groups ({groups.length})</h3>
		{#if groups.length === 0}
			<p class="no-groups">No groups yet. Create your first group to get started!</p>
		{:else}
			<div class="groups-grid">
				{#each groups as group}
					<div class="group-card">
						<div class="group-header">
							<h4>{group.name}</h4>
							<span class="group-id">ID: {group.id}</span>
						</div>
						
						<div class="group-info">
							<div class="admin-info">
								<strong>Admin:</strong> {group.admin.slice(0, 12)}...
							</div>
							<div class="member-count">
								<strong>Members:</strong> {group.members.length}
							</div>
						</div>

						<div class="members-list">
							<h5>Members:</h5>
							{#each group.members as member}
								<div class="member-item">
									<span class="status-indicator status-online"></span>
									{member.slice(0, 12)}...
									{#if member === group.admin}
										<span class="admin-badge">Admin</span>
									{/if}
								</div>
							{/each}
						</div>

						<div class="group-actions">
							<button 
								class="btn btn-secondary"
								on:click={() => openAddMemberModal(group)}
							>
								Add Member
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Create Group Modal -->
{#if showCreateModal}
	<div class="modal" on:click={closeModals}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3 class="modal-title">Create New Group</h3>
			</div>
			
			<div class="form-group">
				<label class="form-label" for="group-id">
					Group ID
				</label>
				<input
					id="group-id"
					type="text"
					class="form-input"
					bind:value={newGroupId}
					placeholder="unique-group-id"
					disabled={loading}
				/>
				<small>A unique identifier for the group</small>
			</div>

			<div class="form-group">
				<label class="form-label" for="group-name">
					Group Name
				</label>
				<input
					id="group-name"
					type="text"
					class="form-input"
					bind:value={newGroupName}
					placeholder="My Awesome Group"
					disabled={loading}
				/>
				<small>A friendly name for the group</small>
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeModals} disabled={loading}>
					Cancel
				</button>
				<button 
					class="btn" 
					on:click={createGroup}
					disabled={!newGroupId.trim() || !newGroupName.trim() || loading}
				>
					{loading ? 'Creating...' : 'Create Group'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Add Member Modal -->
{#if showAddMemberModal && selectedGroup}
	<div class="modal" on:click={closeModals}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h3 class="modal-title">Add Member to {selectedGroup.name}</h3>
			</div>
			
			<div class="form-group">
				<label class="form-label" for="member-id">
					Member Peer ID
				</label>
				<input
					id="member-id"
					type="text"
					class="form-input"
					bind:value={newMemberId}
					placeholder="12D3KooW..."
					disabled={loading}
				/>
				<small>The peer ID of the member you want to add</small>
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" on:click={closeModals} disabled={loading}>
					Cancel
				</button>
				<button 
					class="btn" 
					on:click={addMember}
					disabled={!newMemberId.trim() || loading}
				>
					{loading ? 'Adding...' : 'Add Member'}
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

	.groups-section h3 {
		margin-bottom: 15px;
		color: #333;
	}

	.no-groups {
		text-align: center;
		color: #6c757d;
		font-style: italic;
		padding: 40px 20px;
	}

	.groups-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 20px;
	}

	.group-card {
		background: #f8f9fa;
		border: 1px solid #ddd;
		border-radius: 8px;
		padding: 20px;
	}

	.group-header {
		margin-bottom: 15px;
	}

	.group-header h4 {
		margin: 0 0 5px 0;
		color: #333;
	}

	.group-id {
		font-size: 12px;
		color: #6c757d;
	}

	.group-info {
		margin-bottom: 15px;
		font-size: 14px;
	}

	.group-info div {
		margin-bottom: 5px;
	}

	.members-list {
		margin-bottom: 15px;
	}

	.members-list h5 {
		margin: 0 0 10px 0;
		font-size: 14px;
		color: #333;
	}

	.member-item {
		display: flex;
		align-items: center;
		padding: 5px 0;
		font-size: 12px;
		gap: 8px;
	}

	.admin-badge {
		background: #007bff;
		color: white;
		padding: 2px 6px;
		border-radius: 3px;
		font-size: 10px;
		margin-left: auto;
	}

	.group-actions {
		display: flex;
		gap: 10px;
	}

	h2 {
		margin: 0;
	}

	small {
		color: #6c757d;
		font-size: 12px;
	}
</style>

