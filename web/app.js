// API Configuration
const API_BASE_URL = 'http://localhost:8080';

// Application State
let currentSchedule = null;
let isEditMode = false;

// DOM Elements
const listView = document.getElementById('list-view');
const formView = document.getElementById('form-view');
const detailView = document.getElementById('detail-view');
const scheduleList = document.getElementById('schedule-list');
const scheduleDetail = document.getElementById('schedule-detail');
const form = document.getElementById('schedule-form');
const formTitle = document.getElementById('form-title');
const notification = document.getElementById('notification');
const loading = document.getElementById('loading');

// Initialize application
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    loadSchedules();
});

// Event Listeners
function setupEventListeners() {
    // Toolbar buttons
    document.getElementById('btn-create').addEventListener('click', showCreateForm);
    document.getElementById('btn-refresh').addEventListener('click', () => loadSchedules());
    document.getElementById('btn-cancel').addEventListener('click', showListView);
    document.getElementById('btn-back').addEventListener('click', showListView);

    // Detail view buttons
    document.getElementById('btn-edit').addEventListener('click', showEditForm);
    document.getElementById('btn-approve').addEventListener('click', () => approveSchedule(currentSchedule.id));
    document.getElementById('btn-deny').addEventListener('click', () => denySchedule(currentSchedule.id));
    document.getElementById('btn-delete').addEventListener('click', deleteSchedule);

    // Form submit
    form.addEventListener('submit', handleFormSubmit);

    // Filters
    document.getElementById('filter-owner').addEventListener('input', debounce(loadSchedules, 500));
    document.getElementById('filter-status').addEventListener('change', loadSchedules);
    document.getElementById('filter-env').addEventListener('change', loadSchedules);

    // Event delegation for clickable status badges
    document.getElementById('schedule-list').addEventListener('click', (e) => {
        const statusBadge = e.target.closest('.status-clickable');
        if (statusBadge) {
            e.stopPropagation();
            const scheduleId = statusBadge.dataset.scheduleId;
            const currentStatus = statusBadge.dataset.status;
            showStatusEditor(statusBadge, scheduleId, currentStatus);
        }
    });

    // Note: initializeFormTagInputs() will be called after TagInput class is loaded
}

// Initialize tag input components
let ownerTagInput;
let environmentTagInput;

function initializeFormTagInputs() {
    // Initialize owner tag input
    const ownerContainer = document.getElementById('owner-tag-input');
    if (ownerContainer) {
        ownerTagInput = new TagInput(ownerContainer, {
            placeholder: 'Enter owners separated by ;',
            parseDelimiter: ';',
            validate: (value) => {
                const trimmed = value.trim();
                if (trimmed.length === 0) return false;
                if (trimmed.length > 255) return false;
                return true;
            }
        });
    }

    // Initialize environment tag input
    const envContainer = document.getElementById('environment-tag-input');
    const envOptions = document.getElementById('environment-options');
    if (envContainer && envOptions) {
        environmentTagInput = new EnvironmentTagInput(envContainer, envOptions);
    }
}

// API Calls
async function apiCall(endpoint, options = {}) {
    try {
        showLoading();

        // Get auth token from localStorage
        const token = localStorage.getItem('auth_token');

        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                // Include Authorization header if token exists
                ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
                ...options.headers,
            },
            ...options,
        });

        if (!response.ok) {
            // Handle 401 Unauthorized - redirect to login
            if (response.status === 401) {
                console.log('Unauthorized - redirecting to login');
                localStorage.removeItem('auth_token');
                localStorage.removeItem('user_email');
                localStorage.removeItem('user_name');
                localStorage.removeItem('user_role');
                window.location.href = '/auth/google/login';
                return;
            }

            const error = await response.json();
            throw new Error(error.message || `HTTP error! status: ${response.status}`);
        }

        if (response.status === 204) {
            return null;
        }

        return await response.json();
    } finally {
        hideLoading();
    }
}

async function loadSchedules() {
    const owner = document.getElementById('filter-owner').value;
    const status = document.getElementById('filter-status').value;
    const environment = document.getElementById('filter-env').value;

    const params = new URLSearchParams();
    if (owner) params.append('owner', owner);
    if (status) params.append('status', status);
    if (environment) params.append('environment', environment);

    const query = params.toString() ? `?${params.toString()}` : '';

    try {
        const schedules = await apiCall(`/schedules${query}`);
        displaySchedules(schedules);
    } catch (error) {
        showNotification('Failed to load schedules: ' + error.message, 'error');
    }
}

async function getSchedule(id) {
    try {
        return await apiCall(`/schedules/${id}`);
    } catch (error) {
        showNotification('Failed to load schedule: ' + error.message, 'error');
        return null;
    }
}

async function createSchedule(data) {
    try {
        const schedule = await apiCall('/schedules', {
            method: 'POST',
            body: JSON.stringify(data),
        });
        showNotification('Schedule created successfully', 'success');
        return schedule;
    } catch (error) {
        showNotification('Failed to create schedule: ' + error.message, 'error');
        throw error;
    }
}

async function updateSchedule(id, data) {
    try {
        const schedule = await apiCall(`/schedules/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
        showNotification('Schedule updated successfully', 'success');
        return schedule;
    } catch (error) {
        showNotification('Failed to update schedule: ' + error.message, 'error');
        throw error;
    }
}

async function approveSchedule(id) {
    try {
        const schedule = await apiCall(`/schedules/${id}/approve`, {
            method: 'POST',
        });
        showNotification('Schedule approved successfully', 'success');
        showDetail(schedule);
    } catch (error) {
        showNotification('Failed to approve schedule: ' + error.message, 'error');
    }
}

async function denySchedule(id) {
    try {
        const schedule = await apiCall(`/schedules/${id}/deny`, {
            method: 'POST',
        });
        showNotification('Schedule denied successfully', 'success');
        showDetail(schedule);
    } catch (error) {
        showNotification('Failed to deny schedule: ' + error.message, 'error');
    }
}

async function deleteSchedule() {
    if (!currentSchedule) return;

    if (!confirm(`Are you sure you want to delete schedule ${currentSchedule.id}?`)) {
        return;
    }

    try {
        await apiCall(`/schedules/${currentSchedule.id}`, {
            method: 'DELETE',
        });
        showNotification('Schedule deleted successfully', 'success');
        showListView();
        loadSchedules();
    } catch (error) {
        showNotification('Failed to delete schedule: ' + error.message, 'error');
    }
}

// Helper Functions for Table Display
function getStatusIcon(status) {
    const icons = {
        'approved': '✓',
        'created': '●',
        'denied': '✗'
    };
    const iconClass = `status-icon status-icon--${status}`;
    return `<span class="${iconClass}">${icons[status] || ''}</span>`;
}

function getStatusBadge(status) {
    const statusText = status || 'unknown';
    const badgeClass = `badge badge--status-${statusText}`;
    return `<span class="${badgeClass}">${statusText}</span>`;
}

function getEnvironmentBadge(environment) {
    const envText = environment || 'unknown';
    const badgeClass = `badge badge--env-${envText}`;
    return `<span class="${badgeClass}">${envText}</span>`;
}

function getEnvironmentBadges(environments) {
    if (!environments || environments.length === 0) return '<span class="badge badge--env-unknown">unknown</span>';
    return environments.map(env => {
        const badgeClass = `badge env-${env}`;
        return `<span class="${badgeClass}">${env}</span>`;
    }).join(' ');
}

function getOwnersDisplay(owners, maxDisplay = 3) {
    if (!owners || owners.length === 0) return 'N/A';

    const escaped = owners.map(o => escapeHtml(o));

    if (owners.length <= maxDisplay) {
        return escaped.join(', ');
    }

    const displayed = escaped.slice(0, maxDisplay).join(', ');
    const remaining = owners.length - maxDisplay;
    const allOwners = escaped.join(', ');

    return `${displayed} <span class="owners-more" title="${allOwners}">+${remaining} more</span>`;
}

function getClickableStatusBadge(status, scheduleId) {
    const badgeClass = `badge badge-${status} status-clickable`;
    return `<span class="${badgeClass}" data-schedule-id="${scheduleId}" data-status="${status}">${status}</span>`;
}

// Display Functions
function displaySchedules(schedules) {
    if (!schedules || schedules.length === 0) {
        scheduleList.innerHTML = '<tr><td colspan="6" style="text-align: center; color: #95a5a6; padding: 40px;">No schedules found</td></tr>';
        return;
    }

    scheduleList.innerHTML = schedules.map(schedule => `
        <tr onclick="loadAndShowDetail('${schedule.id}')">
            <td>${schedule.id.substring(0, 8)}</td>
            <td>${escapeHtml(schedule.serviceName)}</td>
            <td>${getEnvironmentBadges(schedule.environments)}</td>
            <td>${formatDateTime(schedule.scheduledAt)}</td>
            <td>${getOwnersDisplay(schedule.owners)}</td>
            <td onclick="event.stopPropagation()">${getStatusIcon(schedule.status)}${getClickableStatusBadge(schedule.status, schedule.id)}</td>
            <td onclick="event.stopPropagation()">
                <button class="btn-icon" onclick="quickAssignGroups('${schedule.id}')" title="Add to group">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                        <path d="M8 2a.5.5 0 0 1 .5.5v5h5a.5.5 0 0 1 0 1h-5v5a.5.5 0 0 1-1 0v-5h-5a.5.5 0 0 1 0-1h5v-5A.5.5 0 0 1 8 2z"/>
                        <circle cx="8" cy="8" r="6.5" stroke="currentColor" fill="none" stroke-width="0.5"/>
                    </svg>
                </button>
            </td>
        </tr>
    `).join('');
}

async function loadAndShowDetail(id) {
    const schedule = await getSchedule(id);
    if (schedule) {
        showDetail(schedule);
    }
}

// Quick assign groups from list view
async function quickAssignGroups(scheduleId) {
    try {
        // Fetch schedule and groups in parallel
        const currentUser = getCurrentUser();
        const [schedule, groupsResponse] = await Promise.all([
            getSchedule(scheduleId),
            fetch(`${API_BASE_URL}/groups?owner=${encodeURIComponent(currentUser)}`, {
                headers: { 'Authorization': `Bearer ${getToken()}` }
            })
        ]);

        if (!schedule) {
            showNotification('Failed to load schedule', 'error');
            return;
        }

        // Check if groups request succeeded
        if (!groupsResponse.ok) {
            const errorText = await groupsResponse.text();
            console.error('Failed to fetch groups:', errorText);
            showNotification('Failed to load groups: ' + (groupsResponse.status === 401 ? 'Not authenticated' : errorText.substring(0, 100)), 'error');
            return;
        }

        const groups = await groupsResponse.json();

        // Get currently assigned group IDs
        const assignedGroupIds = (schedule.groups || []).map(g => g.id);

        // Create modal HTML
        const modalHtml = `
            <div class="quick-assign-overlay" id="quick-assign-overlay" onclick="closeQuickAssign(event)">
                <div class="quick-assign-modal" onclick="event.stopPropagation()">
                    <div class="quick-assign-header">
                        <h3>Add to Groups</h3>
                        <button onclick="closeQuickAssign()" class="modal-close">&times;</button>
                    </div>
                    <div class="quick-assign-body">
                        <p style="margin-bottom: 12px; color: #6b7280; font-size: 14px;">
                            <strong>${escapeHtml(schedule.serviceName)}</strong> - ${formatDateTime(schedule.scheduledAt)}
                        </p>
                        ${groups.length === 0
                            ? '<p style="color: #6b7280;">No groups available. Create a group first.</p>'
                            : `<div class="group-quick-list">
                                ${groups.map(group => {
                                    const isAssigned = assignedGroupIds.includes(group.id);
                                    return `
                                        <label class="group-quick-item ${isAssigned ? 'assigned' : ''}" data-group-id="${group.id}">
                                            <input type="checkbox"
                                                ${isAssigned ? 'checked disabled' : ''}
                                                onchange="handleQuickGroupToggle('${scheduleId}', '${group.id}', this.checked)">
                                            <span class="group-quick-name">
                                                ${group.isFavorite ? '<span class="favorite-indicator">★</span> ' : ''}
                                                ${escapeHtml(group.name)}
                                            </span>
                                            ${isAssigned ? '<span class="assigned-badge">✓ Assigned</span>' : ''}
                                        </label>
                                    `;
                                }).join('')}
                            </div>`
                        }
                    </div>
                </div>
            </div>
        `;

        // Add to DOM
        const overlay = document.createElement('div');
        overlay.innerHTML = modalHtml;
        document.body.appendChild(overlay.firstElementChild);

    } catch (error) {
        console.error('Quick assign error:', error);
        showNotification('Failed to load groups: ' + error.message, 'error');
    }
}

function closeQuickAssign(event) {
    if (event && event.target.id !== 'quick-assign-overlay' && !event.target.classList.contains('modal-close')) {
        return;
    }
    const overlay = document.getElementById('quick-assign-overlay');
    if (overlay) {
        overlay.remove();
    }
}

async function handleQuickGroupToggle(scheduleId, groupId, isChecked) {
    try {
        if (isChecked) {
            // Assign to group
            await fetch(`${API_BASE_URL}/schedules/${scheduleId}/groups`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${getToken()}`
                },
                body: JSON.stringify({ groupIds: [groupId], assignedBy: getCurrentUser() })
            });

            // Update UI - add "assigned" class and check mark
            const label = document.querySelector(`label[data-group-id="${groupId}"]`);
            if (label) {
                label.classList.add('assigned');
                const checkbox = label.querySelector('input');
                checkbox.disabled = true;
                const nameSpan = label.querySelector('.group-quick-name');
                if (!label.querySelector('.assigned-badge')) {
                    nameSpan.insertAdjacentHTML('afterend', '<span class="assigned-badge">✓ Assigned</span>');
                }
            }

            showNotification('Added to group successfully', 'success');

            // Reload schedules to update the list
            setTimeout(() => loadSchedules(), 500);
        }
    } catch (error) {
        console.error('Quick assign toggle error:', error);
        showNotification('Failed to assign group: ' + error.message, 'error');
        // Revert checkbox
        event.target.checked = !isChecked;
    }
}

function showDetail(schedule) {
    currentSchedule = schedule;

    // Update button visibility based on status
    const btnApprove = document.getElementById('btn-approve');
    const btnDeny = document.getElementById('btn-deny');
    const btnEdit = document.getElementById('btn-edit');

    if (schedule.status === 'created') {
        btnApprove.style.display = 'inline-block';
        btnDeny.style.display = 'inline-block';
    } else {
        btnApprove.style.display = 'none';
        btnDeny.style.display = 'none';
    }

    scheduleDetail.innerHTML = `
        <div class="detail-section">
            <div class="detail-label">ID</div>
            <div class="detail-value detail-value-mono">${schedule.id}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Service Name</div>
            <div class="detail-value">${escapeHtml(schedule.serviceName)}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Environments</div>
            <div class="detail-value">${getEnvironmentBadges(schedule.environments)}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Scheduled Time</div>
            <div class="detail-value">${formatDateTime(schedule.scheduledAt)}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Owners</div>
            <div class="detail-value">${schedule.owners ? schedule.owners.map(o => escapeHtml(o)).join(', ') : 'N/A'}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Status</div>
            <div class="detail-value"><span class="status-badge status-${schedule.status}">${schedule.status}</span></div>
        </div>
        ${schedule.description ? `
        <div class="detail-section">
            <div class="detail-label">Description</div>
            <div class="detail-value">${escapeHtml(schedule.description)}</div>
        </div>
        ` : ''}
        ${schedule.rollbackPlan ? `
        <div class="detail-section">
            <div class="detail-label">Rollback Plan</div>
            <div class="detail-value detail-value-multiline">${escapeHtml(schedule.rollbackPlan)}</div>
        </div>
        ` : ''}
        <div class="detail-section">
            <div class="detail-label">Created At</div>
            <div class="detail-value">${formatDateTime(schedule.createdAt)}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Updated At</div>
            <div class="detail-value">${formatDateTime(schedule.updatedAt)}</div>
        </div>
    `;

    showView('detail');
}

// Form Handlers
function showCreateForm() {
    isEditMode = false;
    currentSchedule = null;
    formTitle.textContent = 'Create Schedule';
    form.reset();

    // Clear tag inputs
    if (ownerTagInput) ownerTagInput.clear();
    if (environmentTagInput) environmentTagInput.clear();

    showView('form');
}

function showEditForm() {
    if (!currentSchedule) return;

    isEditMode = true;
    formTitle.textContent = 'Edit Schedule';

    // Populate form
    document.getElementById('scheduled-at').value = toDateTimeLocalString(currentSchedule.scheduledAt);
    document.getElementById('service').value = currentSchedule.serviceName;
    document.getElementById('description').value = currentSchedule.description || '';
    document.getElementById('rollback-plan').value = currentSchedule.rollbackPlan || '';

    // Populate tag inputs
    if (ownerTagInput) {
        ownerTagInput.clear();
        if (currentSchedule.owners) {
            currentSchedule.owners.forEach(owner => ownerTagInput.addTag(owner));
        }
    }

    if (environmentTagInput) {
        environmentTagInput.clear();
        if (currentSchedule.environments) {
            currentSchedule.environments.forEach(env => environmentTagInput.addTag(env));
        }
    }

    showView('form');
}

async function handleFormSubmit(e) {
    e.preventDefault();

    // Get owners and environments from tag inputs
    const owners = ownerTagInput.getTags();
    const environments = environmentTagInput.getEnvironments();

    // Validate at least one owner and environment
    if (owners.length === 0) {
        showNotification('At least one owner is required', 'error');
        return;
    }
    if (environments.length === 0) {
        showNotification('At least one environment is required', 'error');
        return;
    }

    const scheduledAt = new Date(document.getElementById('scheduled-at').value).toISOString();
    const data = {
        scheduledAt,
        serviceName: document.getElementById('service').value,
        environments: environments,
        owners: owners,
    };

    const description = document.getElementById('description').value;
    if (description) {
        data.description = description;
    }

    const rollbackPlan = document.getElementById('rollback-plan').value;
    if (rollbackPlan) {
        data.rollbackPlan = rollbackPlan;
    }

    try {
        let schedule;
        if (isEditMode) {
            schedule = await updateSchedule(currentSchedule.id, data);
        } else {
            schedule = await createSchedule(data);
        }

        showDetail(schedule);
    } catch (error) {
        // Error already shown in API call
    }
}

// View Management
function showView(view) {
    listView.classList.add('hidden');
    formView.classList.add('hidden');
    detailView.classList.add('hidden');

    if (view === 'list') {
        listView.classList.remove('hidden');
    } else if (view === 'form') {
        formView.classList.remove('hidden');
    } else if (view === 'detail') {
        detailView.classList.remove('hidden');
    }
}

function showListView() {
    showView('list');
    loadSchedules();
}

// UI Helpers
function showNotification(message, type = 'success') {
    notification.textContent = message;
    notification.className = `notification ${type}`;
    notification.classList.remove('hidden');

    setTimeout(() => {
        notification.classList.add('hidden');
    }, 5000);
}

function showLoading() {
    loading.classList.remove('hidden');
}

function hideLoading() {
    loading.classList.add('hidden');
}

// Utility Functions
function formatDateTime(dateString) {
    return new Date(dateString).toLocaleString();
}

function toDateTimeLocalString(dateString) {
    const date = new Date(dateString);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showToast(message, type = 'success') {
    showNotification(message, type);
}

function showStatusEditor(badgeElement, scheduleId, currentStatus) {
    const editor = new InlineStatusEditor(scheduleId, currentStatus, async (id, newStatus) => {
        // Update via API
        const data = { status: newStatus };
        await apiCall(`/schedules/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
        // Reload the schedule list to reflect the change
        await loadSchedules();
    });
    editor.show(badgeElement);
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// ===================================
// GROUP MANAGEMENT FUNCTIONALITY
// ===================================

// Group API Calls
async function createGroup(name, description, owner) {
    return await apiCall('/groups', {
        method: 'POST',
        body: JSON.stringify({ name, description, owner })
    });
}

async function fetchGroups(owner) {
    const params = new URLSearchParams();
    if (owner) params.append('owner', owner);
    return await apiCall(`/groups?${params.toString()}`);
}

async function getGroup(id) {
    return await apiCall(`/groups/${id}`);
}

async function updateGroup(id, name, description) {
    return await apiCall(`/groups/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ name, description })
    });
}

async function deleteGroup(id) {
    return await apiCall(`/groups/${id}`, {
        method: 'DELETE'
    });
}

async function assignScheduleToGroups(scheduleId, groupIds, assignedBy = '') {
    return await apiCall(`/schedules/${scheduleId}/groups`, {
        method: 'POST',
        body: JSON.stringify({ groupIds, assignedBy })
    });
}

async function unassignScheduleFromGroup(scheduleId, groupId) {
    return await apiCall(`/schedules/${scheduleId}/groups/${groupId}`, {
        method: 'DELETE'
    });
}

async function getGroupsForSchedule(scheduleId) {
    return await apiCall(`/schedules/${scheduleId}/groups`);
}

async function getSchedulesInGroup(groupId) {
    return await apiCall(`/groups/${groupId}/schedules`);
}

async function bulkAssignSchedules(groupId, scheduleIds, assignedBy = '') {
    return await apiCall(`/groups/${groupId}/schedules`, {
        method: 'POST',
        body: JSON.stringify({ scheduleIds, assignedBy })
    });
}

async function favoriteGroup(groupId) {
    return await apiCall(`/groups/${groupId}/favorite`, {
        method: 'POST'
    });
}

async function unfavoriteGroup(groupId) {
    return await apiCall(`/groups/${groupId}/favorite`, {
        method: 'DELETE'
    });
}

// ===================================
// GROUP CONTROLLER
// ===================================

class GroupController {
    constructor() {
        this.groups = [];
        this.currentGroupId = null;
        this.initializeElements();
        this.setupEventListeners();
    }

    initializeElements() {
        this.groupModal = document.getElementById('group-modal');
        this.groupForm = document.getElementById('group-form');
        this.groupsListModal = document.getElementById('groups-list-modal');
        this.assignGroupModal = document.getElementById('assign-group-modal');
    }

    setupEventListeners() {
        // Create Group button
        document.getElementById('btn-create-group')?.addEventListener('click', () => this.showCreateGroupModal());
        document.getElementById('btn-manage-groups')?.addEventListener('click', () => this.showGroupsListModal());
        
        // Group form
        this.groupForm.addEventListener('submit', (e) => this.handleGroupFormSubmit(e));
        document.getElementById('group-form-cancel')?.addEventListener('click', () => this.hideGroupModal());
        document.getElementById('group-modal-close')?.addEventListener('click', () => this.hideGroupModal());
        
        // Groups list modal
        document.getElementById('groups-list-close')?.addEventListener('click', () => this.hideGroupsListModal());
        
        // Assign group modal
        document.getElementById('btn-assign-group')?.addEventListener('click', () => this.showAssignGroupModal());
        document.getElementById('assign-group-close')?.addEventListener('click', () => this.hideAssignGroupModal());
        document.getElementById('btn-cancel-assignment')?.addEventListener('click', () => this.hideAssignGroupModal());
        document.getElementById('btn-save-assignment')?.addEventListener('click', () => this.handleSaveAssignment());
    }

    async showCreateGroupModal() {
        this.currentGroupId = null;
        document.getElementById('group-modal-title').textContent = 'Create Group';
        document.getElementById('group-name').value = '';
        document.getElementById('group-description').value = '';
        document.getElementById('group-id').value = '';
        this.groupModal.classList.remove('hidden');
    }

    async showEditGroupModal(group) {
        this.currentGroupId = group.id;
        document.getElementById('group-modal-title').textContent = 'Edit Group';
        document.getElementById('group-name').value = group.name;
        document.getElementById('group-description').value = group.description || '';
        document.getElementById('group-id').value = group.id;
        this.groupModal.classList.remove('hidden');
    }

    hideGroupModal() {
        this.groupModal.classList.add('hidden');
        this.groupForm.reset();
    }

    async showGroupsListModal() {
        try {
            const owner = document.getElementById('filter-owner').value || getCurrentUser();
            this.groups = await fetchGroups(owner);
            this.renderGroupsList();
            this.groupsListModal.classList.remove('hidden');
        } catch (error) {
            showNotification('Failed to load groups: ' + error.message, 'error');
        }
    }

    hideGroupsListModal() {
        this.groupsListModal.classList.add('hidden');
    }

    async showAssignGroupModal() {
        if (!currentSchedule) return;
        
        try {
            const owner = getCurrentUser();
            this.groups = await fetchGroups(owner);
            
            // Get currently assigned groups
            const assignedGroups = currentSchedule.groups || [];
            const assignedIds = assignedGroups.map(g => g.id);
            
            this.renderGroupCheckboxList(assignedIds);
            this.assignGroupModal.classList.remove('hidden');
        } catch (error) {
            showNotification('Failed to load groups: ' + error.message, 'error');
        }
    }

    hideAssignGroupModal() {
        this.assignGroupModal.classList.add('hidden');
    }

    renderGroupsList() {
        const tbody = document.getElementById('groups-list');
        tbody.innerHTML = '';

        if (this.groups.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" style="text-align: center; padding: 20px;">No groups found. Create your first group!</td></tr>';
            return;
        }

        this.groups.forEach(group => {
            const row = document.createElement('tr');
            const starClass = group.isFavorite ? 'favorite-star favorite-star--filled' : 'favorite-star';
            const starIcon = group.isFavorite ? '★' : '☆';

            row.innerHTML = `
                <td style="text-align: center; width: 40px;">
                    <span class="${starClass}" data-group-id="${group.id}" style="cursor: pointer; font-size: 18px;" title="${group.isFavorite ? 'Remove from favorites' : 'Add to favorites'}">${starIcon}</span>
                </td>
                <td><strong>${escapeHtml(group.name)}</strong></td>
                <td>${escapeHtml(group.description || '')}</td>
                <td><span class="badge">0</span></td>
                <td class="group-actions">
                    <button class="btn btn-sm btn-secondary" onclick="groupController.showEditGroupModal(${JSON.stringify(group).replace(/"/g, '&quot;')})">Edit</button>
                    <button class="btn btn-sm btn-danger" onclick="groupController.handleDeleteGroup('${group.id}')">Delete</button>
                </td>
            `;
            tbody.appendChild(row);
        });

        // Add click handlers for favorite stars in the table
        tbody.querySelectorAll('.favorite-star').forEach(star => {
            star.addEventListener('click', (e) => {
                e.stopPropagation();
                const groupId = star.dataset.groupId;
                this.toggleFavoriteInList(groupId);
            });
        });
    }

    renderGroupCheckboxList(selectedIds = []) {
        const container = document.getElementById('group-checkbox-list');
        container.innerHTML = '';
        
        if (this.groups.length === 0) {
            container.innerHTML = '<p style="padding: 20px; text-align: center;">No groups available. Create a group first.</p>';
            return;
        }
        
        this.groups.forEach(group => {
            const div = document.createElement('div');
            div.className = 'checkbox-item';
            
            const checked = selectedIds.includes(group.id) ? 'checked' : '';
            div.innerHTML = `
                <input type="checkbox" id="group-${group.id}" value="${group.id}" ${checked}>
                <label for="group-${group.id}">
                    <div><strong>${escapeHtml(group.name)}</strong></div>
                    ${group.description ? `<div class="checkbox-item__description">${escapeHtml(group.description)}</div>` : ''}
                </label>
            `;
            container.appendChild(div);
        });
    }

    async handleGroupFormSubmit(e) {
        e.preventDefault();
        
        const name = document.getElementById('group-name').value.trim();
        const description = document.getElementById('group-description').value.trim();
        const owner = getCurrentUser();
        
        try {
            if (this.currentGroupId) {
                await updateGroup(this.currentGroupId, name, description);
                showNotification('Group updated successfully', 'success');
            } else {
                await createGroup(name, description, owner);
                showNotification('Group created successfully', 'success');
            }
            
            this.hideGroupModal();
            
            // Refresh tabs and groups list if open
            if (window.tabController) {
                await window.tabController.loadTabs();
            }
            if (!this.groupsListModal.classList.contains('hidden')) {
                await this.showGroupsListModal();
            }
        } catch (error) {
            showNotification('Failed to save group: ' + error.message, 'error');
        }
    }

    async handleDeleteGroup(id) {
        if (!confirm('Are you sure you want to delete this group? This will unassign it from all schedules.')) {
            return;
        }

        try {
            await deleteGroup(id);
            showNotification('Group deleted successfully', 'success');

            // Refresh tabs and groups list
            if (window.tabController) {
                await window.tabController.loadTabs();
            }
            await this.showGroupsListModal();
        } catch (error) {
            showNotification('Failed to delete group: ' + error.message, 'error');
        }
    }

    async toggleFavoriteInList(groupId) {
        // Find the group
        const group = this.groups.find(g => g.id === groupId);
        if (!group) return;

        // Store original state for rollback on error
        const originalState = group.isFavorite;

        try {
            // Optimistic UI update
            group.isFavorite = !group.isFavorite;
            this.renderGroupsList();

            // Call API
            if (group.isFavorite) {
                await favoriteGroup(groupId);
            } else {
                await unfavoriteGroup(groupId);
            }

            // Refresh tabs to show updated star
            if (window.tabController) {
                await window.tabController.loadTabs();
            }

        } catch (error) {
            console.error('Failed to toggle favorite:', error);
            // Revert optimistic update
            group.isFavorite = originalState;
            this.renderGroupsList();
            showNotification('Failed to update favorite: ' + error.message, 'error');
        }
    }

    async handleSaveAssignment() {
        if (!currentSchedule) return;
        
        const checkboxes = document.querySelectorAll('#group-checkbox-list input[type="checkbox"]');
        const selectedGroupIds = Array.from(checkboxes)
            .filter(cb => cb.checked)
            .map(cb => cb.value);
        
        try {
            await assignScheduleToGroups(currentSchedule.id, selectedGroupIds, getCurrentUser());
            showNotification('Groups assigned successfully', 'success');
            this.hideAssignGroupModal();
            
            // Refresh schedule details to show updated groups
            const schedule = await getSchedule(currentSchedule.id);
            if (schedule) {
                showDetail(schedule);
            }
        } catch (error) {
            showNotification('Failed to assign groups: ' + error.message, 'error');
        }
    }

    renderGroupBadges(groups, scheduleId) {
        const container = document.getElementById('schedule-groups');
        if (!container) return;
        
        container.innerHTML = '';
        
        if (!groups || groups.length === 0) {
            container.innerHTML = '<p style="color: #6b7280; font-size: 14px;">No groups assigned</p>';
            return;
        }
        
        groups.forEach(group => {
            const badge = document.createElement('div');
            badge.className = 'group-badge';
            badge.innerHTML = `
                <span>${escapeHtml(group.name)}</span>
                <button class="group-badge__remove" onclick="groupController.removeGroupFromSchedule('${scheduleId}', '${group.id}')" title="Remove">×</button>
            `;
            container.appendChild(badge);
        });
    }

    async removeGroupFromSchedule(scheduleId, groupId) {
        try {
            await unassignScheduleFromGroup(scheduleId, groupId);
            showNotification('Group removed from schedule', 'success');
            
            // Refresh schedule details
            const schedule = await getSchedule(scheduleId);
            if (schedule) {
                showDetail(schedule);
            }
        } catch (error) {
            showNotification('Failed to remove group: ' + error.message, 'error');
        }
    }
}

// ===================================
// TAB CONTROLLER
// ===================================

class TabController {
    constructor() {
        this.tabs = [];
        this.activeTabId = this.loadActiveTab() || 'all';
        this.groups = [];
        this.initializeElements();
    }

    initializeElements() {
        this.tabList = document.getElementById('tab-list');
    }

    async init() {
        await this.loadTabs();
    }

    async loadTabs() {
        try {
            const owner = getCurrentUser();
            this.groups = await fetchGroups(owner);
            this.buildTabs();
            this.renderTabs();
        } catch (error) {
            console.error('Failed to load tabs:', error);
            this.tabs = [{ id: 'all', label: 'All Schedules', count: 0 }];
            this.renderTabs();
        }
    }

    buildTabs() {
        this.tabs = [
            { id: 'all', label: 'All Schedules', count: 0 },
            { id: 'ungrouped', label: 'Ungrouped', count: 0 }
        ];
        
        // Add group tabs
        this.groups.forEach(group => {
            this.tabs.push({
                id: group.id,
                label: group.name,
                count: 0,
                isGroup: true,
                isFavorite: group.isFavorite || false
            });
        });
    }

    renderTabs() {
        this.tabList.innerHTML = '';

        this.tabs.forEach(tab => {
            const li = document.createElement('li');
            li.setAttribute('role', 'presentation');

            const button = document.createElement('button');
            button.className = 'tab-nav__tab';
            button.setAttribute('role', 'tab');
            button.setAttribute('aria-selected', tab.id === this.activeTabId ? 'true' : 'false');

            if (tab.id === this.activeTabId) {
                button.classList.add('tab-nav__tab--active');
            }

            // Build tab content with optional favorite star
            let starHtml = '';
            if (tab.isGroup) {
                const starClass = tab.isFavorite ? 'favorite-star favorite-star--filled' : 'favorite-star';
                const starIcon = tab.isFavorite ? '★' : '☆';
                starHtml = `<span class="${starClass}" data-group-id="${tab.id}" title="${tab.isFavorite ? 'Remove from favorites' : 'Add to favorites'}">${starIcon}</span>`;
            }

            button.innerHTML = `
                ${starHtml}
                <span>${escapeHtml(tab.label)}</span>
                <span class="tab-nav__tab__count">${tab.count}</span>
            `;

            button.addEventListener('click', () => this.switchTab(tab.id));

            li.appendChild(button);
            this.tabList.appendChild(li);
        });

        // Add click handlers for favorite stars
        this.tabList.querySelectorAll('.favorite-star').forEach(star => {
            star.addEventListener('click', (e) => {
                e.stopPropagation(); // Prevent tab switch
                const groupId = star.dataset.groupId;
                this.toggleFavorite(groupId);
            });
        });
    }

    switchTab(tabId) {
        this.activeTabId = tabId;
        this.saveActiveTab(tabId);
        this.renderTabs();
        loadSchedules();
    }

    updateTabCounts(schedules) {
        this.tabs.forEach(tab => {
            if (tab.id === 'all') {
                tab.count = schedules.length;
            } else if (tab.id === 'ungrouped') {
                tab.count = schedules.filter(s => !s.groups || s.groups.length === 0).length;
            } else if (tab.isGroup) {
                tab.count = schedules.filter(s => 
                    s.groups && s.groups.some(g => g.id === tab.id)
                ).length;
            }
        });
        this.renderTabs();
    }

    filterSchedulesByActiveTab(schedules) {
        if (this.activeTabId === 'all') {
            return schedules;
        } else if (this.activeTabId === 'ungrouped') {
            return schedules.filter(s => !s.groups || s.groups.length === 0);
        } else {
            // Filter by group ID
            return schedules.filter(s => 
                s.groups && s.groups.some(g => g.id === this.activeTabId)
            );
        }
    }

    saveActiveTab(tabId) {
        try {
            localStorage.setItem('activeTab', tabId);
        } catch (e) {
            console.error('Failed to save active tab:', e);
        }
    }

    loadActiveTab() {
        try {
            return localStorage.getItem('activeTab');
        } catch (e) {
            console.error('Failed to load active tab:', e);
            return null;
        }
    }

    async toggleFavorite(groupId) {
        // Find the tab
        const tab = this.tabs.find(t => t.id === groupId);
        if (!tab) return;

        // Store original state for rollback on error
        const originalState = tab.isFavorite;

        try {
            // Optimistic UI update
            tab.isFavorite = !tab.isFavorite;
            this.renderTabs();

            // Call API
            if (tab.isFavorite) {
                await favoriteGroup(groupId);
            } else {
                await unfavoriteGroup(groupId);
            }

            // Also update in groups array
            const group = this.groups.find(g => g.id === groupId);
            if (group) {
                group.isFavorite = tab.isFavorite;
            }

            // Re-sort groups (favorites first) and rebuild tabs
            await this.loadTabs();

        } catch (error) {
            console.error('Failed to toggle favorite:', error);
            // Revert optimistic update
            tab.isFavorite = originalState;
            this.renderTabs();
            showNotification('Failed to update favorite: ' + error.message, 'error');
        }
    }
}

// ===================================
// UTILITY FUNCTIONS
// ===================================

function getCurrentUser() {
    // Get from filter or use default
    return document.getElementById('filter-owner').value || 'system';
}

function getToken() {
    return localStorage.getItem('auth_token');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ===================================
// INITIALIZE CONTROLLERS
// ===================================

let groupController;
let tabController;

// Update DOMContentLoaded to initialize controllers
const originalLoad = document.addEventListener('DOMContentLoaded', async () => {
    setupEventListeners();
    
    // Initialize controllers
    groupController = new GroupController();
    window.groupController = groupController; // Make available globally for onclick handlers
    
    tabController = new TabController();
    window.tabController = tabController;
    
    await tabController.init();
    await loadSchedules();
});

// Update loadSchedules to work with tabs
const originalLoadSchedules = loadSchedules;
loadSchedules = async function() {
    const owner = document.getElementById('filter-owner').value;
    const status = document.getElementById('filter-status').value;
    const environment = document.getElementById('filter-env').value;

    const params = new URLSearchParams();
    if (owner) params.append('owner', owner);
    if (status) params.append('status', status);
    if (environment) params.append('environment', environment);
    
    // Add ungrouped filter if ungrouped tab is active
    if (tabController && tabController.activeTabId === 'ungrouped') {
        params.append('ungrouped', 'true');
    }

    const query = params.toString() ? `?${params.toString()}` : '';

    try {
        let schedules = await apiCall(`/schedules${query}`);
        
        // Filter by active tab
        if (tabController) {
            schedules = tabController.filterSchedulesByActiveTab(schedules);
            tabController.updateTabCounts(schedules);
        }
        
        displaySchedules(schedules);
    } catch (error) {
        showNotification('Failed to load schedules: ' + error.message, 'error');
    }
};

// Update showDetail to render group badges
const originalShowScheduleDetail = showDetail;
showDetail = function(schedule) {
    if (originalShowScheduleDetail) {
        originalShowScheduleDetail(schedule);
    }
    
    // Render group badges
    if (groupController) {
        groupController.renderGroupBadges(schedule.groups || [], schedule.id);
    }
};

// ========================================
// Tag Input Components
// ========================================

class TagInput {
    constructor(container, options = {}) {
        this.container = container;
        this.options = {
            placeholder: options.placeholder || 'Enter values separated by ;',
            validate: options.validate || (() => true),
            onValidationError: options.onValidationError || ((error) => alert(error)),
            onChange: options.onChange || (() => {}),
            parseDelimiter: options.parseDelimiter || ';',
            allowDuplicates: options.allowDuplicates || false,
            trim: options.trim !== false
        };
        this.tags = [];
        this.input = null;
        this.init();
    }

    init() {
        this.container.classList.add('tag-input-container');
        this.container.innerHTML = '';
        
        this.input = document.createElement('input');
        this.input.type = 'text';
        this.input.className = 'tag-input';
        this.input.placeholder = this.options.placeholder;
        
        this.input.addEventListener('keydown', (e) => this.handleKeyDown(e));
        this.input.addEventListener('blur', () => this.handleBlur());
        this.input.addEventListener('input', (e) => this.handleInput(e));
        
        this.container.appendChild(this.input);
    }

    handleKeyDown(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            this.addTagFromInput();
        } else if (e.key === 'Backspace' && this.input.value === '' && this.tags.length > 0) {
            this.removeTag(this.tags[this.tags.length - 1]);
        }
    }

    handleBlur() {
        this.addTagFromInput();
    }

    handleInput(e) {
        const value = e.target.value;
        if (value.includes(this.options.parseDelimiter)) {
            const parts = value.split(this.options.parseDelimiter);
            const toProcess = parts.slice(0, -1);
            this.input.value = parts[parts.length - 1];
            
            toProcess.forEach(part => {
                const trimmed = this.options.trim ? part.trim() : part;
                if (trimmed) {
                    this.addTag(trimmed);
                }
            });
        }
    }

    addTagFromInput() {
        const value = this.options.trim ? this.input.value.trim() : this.input.value;
        if (value) {
            this.addTag(value);
            this.input.value = '';
        }
    }

    addTag(value) {
        // Check for duplicates
        if (!this.options.allowDuplicates && this.tags.includes(value)) {
            this.options.onValidationError('Value already added');
            return false;
        }

        // Validate
        const validationResult = this.options.validate(value);
        if (validationResult !== true) {
            this.options.onValidationError(validationResult || 'Invalid value');
            return false;
        }

        this.tags.push(value);
        this.renderTags();
        this.options.onChange(this.tags);
        return true;
    }

    removeTag(value) {
        const index = this.tags.indexOf(value);
        if (index > -1) {
            this.tags.splice(index, 1);
            this.renderTags();
            this.options.onChange(this.tags);
        }
    }

    renderTags() {
        // Remove existing tag elements
        this.container.querySelectorAll('.tag').forEach(el => el.remove());
        
        // Add tag elements
        this.tags.forEach(tag => {
            const tagEl = document.createElement('span');
            tagEl.className = 'tag';
            tagEl.textContent = tag;
            
            const removeBtn = document.createElement('button');
            removeBtn.className = 'tag-remove';
            removeBtn.innerHTML = '&times;';
            removeBtn.onclick = () => this.removeTag(tag);
            
            tagEl.appendChild(removeBtn);
            this.container.insertBefore(tagEl, this.input);
        });
    }

    getTags() {
        return [...this.tags];
    }

    setTags(tags) {
        this.tags = [...tags];
        this.renderTags();
        this.options.onChange(this.tags);
    }

    clear() {
        this.tags = [];
        this.renderTags();
        this.options.onChange(this.tags);
    }

    validate() {
        if (this.tags.length === 0) {
            return 'At least one value is required';
        }
        return true;
    }
}

class EnvironmentTagInput {
    constructor(container, options = {}) {
        this.container = container;
        this.options = {
            availableEnvironments: options.availableEnvironments || ['production', 'staging', 'development'],
            onChange: options.onChange || (() => {}),
            colorMap: options.colorMap || {
                'production': 'env-production',
                'staging': 'env-staging',
                'development': 'env-development'
            }
        };
        this.selectedEnvironments = [];
        this.init();
    }

    init() {
        this.container.classList.add('tag-input-container');
        this.container.innerHTML = '';
        this.render();
    }

    render() {
        this.container.innerHTML = '';
        
        // Render selected environment tags
        this.selectedEnvironments.forEach(env => {
            const tagEl = document.createElement('span');
            tagEl.className = `tag ${this.options.colorMap[env] || ''}`;
            tagEl.textContent = env;
            
            const removeBtn = document.createElement('button');
            removeBtn.className = 'tag-remove';
            removeBtn.innerHTML = '&times;';
            removeBtn.onclick = () => this.removeEnvironment(env);
            
            tagEl.appendChild(removeBtn);
            this.container.appendChild(tagEl);
        });

        // Render available environment options
        const availableEnvs = this.options.availableEnvironments.filter(
            env => !this.selectedEnvironments.includes(env)
        );

        if (availableEnvs.length > 0) {
            const dropdown = document.createElement('select');
            dropdown.className = 'tag-input';
            
            const placeholder = document.createElement('option');
            placeholder.value = '';
            placeholder.textContent = 'Add environment...';
            placeholder.disabled = true;
            placeholder.selected = true;
            dropdown.appendChild(placeholder);
            
            availableEnvs.forEach(env => {
                const option = document.createElement('option');
                option.value = env;
                option.textContent = env;
                dropdown.appendChild(option);
            });
            
            dropdown.onchange = (e) => {
                if (e.target.value) {
                    this.addEnvironment(e.target.value);
                }
            };
            
            this.container.appendChild(dropdown);
        }
    }

    addEnvironment(env) {
        if (!this.selectedEnvironments.includes(env)) {
            this.selectedEnvironments.push(env);
            this.render();
            this.options.onChange(this.selectedEnvironments);
        }
    }

    removeEnvironment(env) {
        if (this.selectedEnvironments.length === 1) {
            alert('At least one environment is required');
            return;
        }
        
        const index = this.selectedEnvironments.indexOf(env);
        if (index > -1) {
            this.selectedEnvironments.splice(index, 1);
            this.render();
            this.options.onChange(this.selectedEnvironments);
        }
    }

    getEnvironments() {
        return [...this.selectedEnvironments];
    }

    setEnvironments(environments) {
        this.selectedEnvironments = [...environments];
        this.render();
        this.options.onChange(this.selectedEnvironments);
    }

    clear() {
        if (this.selectedEnvironments.length > 0) {
            this.selectedEnvironments = [];
            this.render();
            this.options.onChange(this.selectedEnvironments);
        }
    }

    validate() {
        if (this.selectedEnvironments.length === 0) {
            return 'At least one environment is required';
        }
        return true;
    }
}

// ========================================
// Inline Status Editor
// ========================================

class InlineStatusEditor {
    constructor(scheduleId, currentStatus, onStatusChange) {
        this.scheduleId = scheduleId;
        this.currentStatus = currentStatus;
        this.onStatusChange = onStatusChange;
        this.isLoading = false;
        this.availableStatuses = ['created', 'approved', 'denied'];
    }

    show(badgeElement) {
        // Make the badge element keyboard accessible if not already
        if (!badgeElement.hasAttribute('tabindex')) {
            badgeElement.setAttribute('tabindex', '0');
        }

        // Add keyboard event listener
        const keydownHandler = (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                e.stopPropagation();
                this.showDropdown(badgeElement);
            }
        };
        badgeElement.addEventListener('keydown', keydownHandler);
        this.badgeKeydownHandler = keydownHandler;
    }

    createStatusBadge() {
        const container = document.createElement('div');
        container.className = 'status-badge clickable';
        container.tabIndex = 0;

        const badge = document.createElement('span');
        badge.className = `badge badge-${this.currentStatus}`;
        badge.textContent = this.currentStatus;

        container.appendChild(badge);

        container.addEventListener('click', (e) => {
            e.stopPropagation();
            this.showDropdown(container);
        });

        // Keyboard support to open dropdown
        container.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                e.stopPropagation();
                this.showDropdown(container);
            }
        });

        return container;
    }

    showDropdown(container) {
        if (this.isLoading) return;

        // Remove any existing dropdowns
        document.querySelectorAll('.status-dropdown').forEach(el => el.remove());

        const dropdown = document.createElement('div');
        dropdown.className = 'status-dropdown';

        this.availableStatuses.forEach((status, index) => {
            const option = document.createElement('div');
            option.className = 'status-dropdown-option';
            option.innerHTML = `<span class="badge badge-${status}">${status}</span>`;
            option.tabIndex = 0;

            if (status === this.currentStatus) {
                option.style.background = 'var(--color-bg-gray-lighter)';
            }

            option.addEventListener('click', async () => {
                await this.changeStatus(status, container);
                dropdown.remove();
                this.removeKeyboardHandlers();
            });

            // Keyboard navigation for options
            option.addEventListener('keydown', async (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    await this.changeStatus(status, container);
                    dropdown.remove();
                    this.removeKeyboardHandlers();
                }
            });

            dropdown.appendChild(option);
        });

        container.appendChild(dropdown);

        // Focus first option
        const firstOption = dropdown.querySelector('.status-dropdown-option');
        if (firstOption) {
            firstOption.focus();
        }

        // Close dropdown when clicking outside
        setTimeout(() => {
            const closeDropdown = (e) => {
                if (!container.contains(e.target)) {
                    dropdown.remove();
                    this.removeKeyboardHandlers();
                    document.removeEventListener('click', closeDropdown);
                }
            };
            document.addEventListener('click', closeDropdown);
            this.closeDropdownHandler = closeDropdown;
        }, 0);

        // Close dropdown on Escape key
        const handleEscape = (e) => {
            if (e.key === 'Escape') {
                dropdown.remove();
                this.removeKeyboardHandlers();
            }
        };
        document.addEventListener('keydown', handleEscape);
        this.escapeHandler = handleEscape;
    }

    removeKeyboardHandlers() {
        if (this.escapeHandler) {
            document.removeEventListener('keydown', this.escapeHandler);
            this.escapeHandler = null;
        }
        if (this.closeDropdownHandler) {
            document.removeEventListener('click', this.closeDropdownHandler);
            this.closeDropdownHandler = null;
        }
    }

    async changeStatus(newStatus, container) {
        if (newStatus === this.currentStatus || this.isLoading) return;
        
        const previousStatus = this.currentStatus;
        const badge = container.querySelector('.badge');
        
        // Optimistic update
        this.currentStatus = newStatus;
        badge.className = `badge badge-${newStatus} loading`;
        badge.textContent = newStatus;
        this.isLoading = true;
        
        try {
            await this.onStatusChange(this.scheduleId, newStatus);
            // Success
            badge.classList.remove('loading');
            this.isLoading = false;
            showToast('Status updated successfully', 'success');
        } catch (error) {
            // Rollback on error
            this.currentStatus = previousStatus;
            badge.className = `badge badge-${previousStatus}`;
            badge.textContent = previousStatus;
            this.isLoading = false;
            showToast('Failed to update status: ' + error.message, 'error');
        }
    }
}

// Toast notification helper
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    toast.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        padding: 12px 20px;
        background: ${type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : '#3b82f6'};
        color: white;
        border-radius: 4px;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        z-index: 10000;
        animation: slideIn 0.3s ease-out;
    `;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease-out';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Add animations for toast
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(400px); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(400px); opacity: 0; }
    }
`;
document.head.appendChild(style);

// ========================================
// Quick Create Modal
// ========================================

let quickCreateModal;
let quickCreateForm;
let selectedTimeOffset = 0;
let currentUser = null;
let quickCreateGroups = [];
let quickCreateGroupsLoading = false;

// Initialize Quick Create on DOM load
document.addEventListener('DOMContentLoaded', () => {
    initializeQuickCreate();
    loadCurrentUser();
});

function initializeQuickCreate() {
    quickCreateModal = document.getElementById('quick-create-modal');
    quickCreateForm = document.getElementById('quick-create-form');
    
    // Button handlers
    document.getElementById('btn-quick-create')?.addEventListener('click', openQuickCreate);
    document.getElementById('quick-create-close')?.addEventListener('click', closeQuickCreate);
    document.getElementById('btn-quick-cancel')?.addEventListener('click', closeQuickCreate);
    document.getElementById('link-full-form')?.addEventListener('click', (e) => {
        e.preventDefault();
        closeQuickCreate();
        showCreateForm();
    });
    
    // Form submission
    quickCreateForm?.addEventListener('submit', handleQuickCreateSubmit);
    
    // Time offset buttons
    document.querySelectorAll('.time-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            handleTimeOffsetClick(btn);
        });
    });
    
    // Keyboard shortcut (Q key)
    document.addEventListener('keydown', (e) => {
        // Don't trigger if user is typing in an input
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;
        
        if (e.key === 'q' || e.key === 'Q') {
            e.preventDefault();
            openQuickCreate();
        }
    });
    
    // ESC key to close
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !quickCreateModal?.classList.contains('hidden')) {
            closeQuickCreate();
        }
    });
    
    // Click outside to close
    quickCreateModal?.addEventListener('click', (e) => {
        if (e.target === quickCreateModal) {
            closeQuickCreate();
        }
    });
}

async function loadCurrentUser() {
    try {
        const token = localStorage.getItem('auth_token');
        if (!token) {
            console.log('No auth token - redirecting to login');
            window.location.href = '/auth/google/login';
            return;
        }

        const response = await fetch(`${API_BASE_URL}/users/me`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.status === 401) {
            console.log('Unauthorized - redirecting to login');
            localStorage.removeItem('auth_token');
            localStorage.removeItem('user_email');
            localStorage.removeItem('user_name');
            localStorage.removeItem('user_role');
            window.location.href = '/auth/google/login';
            return;
        }

        if (response.ok) {
            currentUser = await response.json();
            document.getElementById('quick-current-user').textContent = currentUser.name || currentUser.email;
        }
    } catch (error) {
        console.error('Failed to load current user:', error);
    }
}

async function loadRecentServices() {
    try {
        const token = localStorage.getItem('auth_token');
        if (!token) return;

        const response = await fetch(`${API_BASE_URL}/services/recent`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            const datalist = document.getElementById('recent-services');
            datalist.innerHTML = '';

            data.services.forEach(service => {
                const option = document.createElement('option');
                option.value = service;
                datalist.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Failed to load recent services:', error);
    }
}

// Task 1.2: Fetch user's groups for Quick Create with favorites-first ordering
async function loadQuickCreateGroups() {
    const loadingEl = document.getElementById('quick-groups-loading');
    const emptyEl = document.getElementById('quick-groups-empty');
    const listEl = document.getElementById('quick-groups-list');

    // Task 1.3: Show loading state
    loadingEl.classList.remove('hidden');
    emptyEl.classList.add('hidden');
    listEl.innerHTML = '';
    quickCreateGroupsLoading = true;

    try {
        const token = localStorage.getItem('auth_token');
        if (!token) {
            quickCreateGroups = [];
            renderQuickCreateGroups();
            return;
        }

        const response = await fetch(`${API_BASE_URL}/groups`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            quickCreateGroups = await response.json();
            // Groups are already ordered with favorites first by the API
        } else {
            quickCreateGroups = [];
        }
    } catch (error) {
        console.error('Failed to load groups for Quick Create:', error);
        quickCreateGroups = [];
    } finally {
        quickCreateGroupsLoading = false;
        loadingEl.classList.add('hidden');
        renderQuickCreateGroups();
    }
}

// Task 1.4: Render checkbox-based multi-select UI with favorite indicators
// Task 1.5: Show empty state when user has no groups
// Task 1.6: Implement keyboard navigation support
function renderQuickCreateGroups() {
    const emptyEl = document.getElementById('quick-groups-empty');
    const listEl = document.getElementById('quick-groups-list');

    listEl.innerHTML = '';

    // Task 1.5: Empty state
    if (!quickCreateGroups || quickCreateGroups.length === 0) {
        emptyEl.classList.remove('hidden');
        return;
    }

    emptyEl.classList.add('hidden');

    // Task 1.4: Render checkboxes with favorite indicators
    quickCreateGroups.forEach((group, index) => {
        const div = document.createElement('div');
        div.className = 'checkbox-item';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.id = `quick-group-${group.id}`;
        checkbox.value = group.id;
        checkbox.name = 'quick-group';

        const label = document.createElement('label');
        label.htmlFor = `quick-group-${group.id}`;

        // Add favorite indicator
        const favoriteIcon = group.isFavorite
            ? '<span class="favorite-indicator" title="Favorite">★</span> '
            : '';

        label.innerHTML = `
            ${favoriteIcon}<strong>${escapeHtml(group.name)}</strong>
            ${group.description ? `<div class="checkbox-item__description">${escapeHtml(group.description)}</div>` : ''}
        `;

        div.appendChild(checkbox);
        div.appendChild(label);

        // Task 1.6: Keyboard navigation support
        checkbox.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                const nextCheckbox = listEl.querySelectorAll('input[type="checkbox"]')[index + 1];
                if (nextCheckbox) nextCheckbox.focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                const prevCheckbox = listEl.querySelectorAll('input[type="checkbox"]')[index - 1];
                if (prevCheckbox) prevCheckbox.focus();
            }
        });

        listEl.appendChild(div);
    });
}

async function openQuickCreate() {
    quickCreateModal.classList.remove('hidden');
    resetQuickCreateForm();
    loadRecentServices();

    // Load groups for selection (Task 1.2)
    await loadQuickCreateGroups();

    // Focus on service input
    setTimeout(() => {
        document.getElementById('quick-service')?.focus();
    }, 100);
}

function closeQuickCreate() {
    quickCreateModal.classList.add('hidden');
    resetQuickCreateForm();
}

function resetQuickCreateForm() {
    quickCreateForm.reset();
    selectedTimeOffset = 0;

    // Reset time buttons
    document.querySelectorAll('.time-btn').forEach(btn => {
        btn.classList.remove('active');
        if (btn.dataset.offset === '0') {
            btn.classList.add('active');
        }
    });

    // Hide custom time input
    document.getElementById('quick-time-custom').classList.add('hidden');

    // Reset checkboxes to default (staging checked)
    document.querySelectorAll('input[name="quick-env"]').forEach(cb => {
        cb.checked = cb.value === 'staging';
    });

    // Reset group checkboxes
    document.querySelectorAll('input[name="quick-group"]').forEach(cb => {
        cb.checked = false;
    });

    // Hide loading state
    document.getElementById('btn-quick-submit').disabled = false;
    document.querySelector('.btn-text').classList.remove('hidden');
    document.querySelector('.btn-loading').classList.add('hidden');
}

function handleTimeOffsetClick(btn) {
    const offset = btn.dataset.offset;
    
    // Update active state
    document.querySelectorAll('.time-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    
    if (offset === 'custom') {
        // Show custom time input
        const customInput = document.getElementById('quick-time-custom');
        customInput.classList.remove('hidden');
        customInput.required = true;
        customInput.focus();
        selectedTimeOffset = null;
    } else {
        // Hide custom time input
        const customInput = document.getElementById('quick-time-custom');
        customInput.classList.add('hidden');
        customInput.required = false;
        selectedTimeOffset = parseInt(offset);
    }
}

async function handleQuickCreateSubmit(e) {
    e.preventDefault();

    // Get form values
    const serviceName = document.getElementById('quick-service').value.trim();
    const selectedEnvs = Array.from(document.querySelectorAll('input[name="quick-env"]:checked'))
        .map(cb => cb.value);

    // Task 2.1: Get selected group IDs
    const selectedGroupIds = Array.from(document.querySelectorAll('input[name="quick-group"]:checked'))
        .map(cb => cb.value);

    // Validation
    if (!serviceName) {
        showToast('Service name is required', 'error');
        return;
    }

    if (selectedEnvs.length === 0) {
        showToast('At least one environment is required', 'error');
        return;
    }

    // Calculate scheduled time
    let scheduledAt;
    if (selectedTimeOffset === null) {
        // Custom time
        const customTime = document.getElementById('quick-time-custom').value;
        if (!customTime) {
            showToast('Please select a time', 'error');
            return;
        }
        scheduledAt = new Date(customTime).toISOString();
    } else {
        // Offset from now
        const now = new Date();
        now.setMinutes(now.getMinutes() + selectedTimeOffset);
        scheduledAt = now.toISOString();
    }

    // Prepare schedule data with smart defaults
    const scheduleData = {
        scheduledAt: scheduledAt,
        serviceName: serviceName,
        environments: selectedEnvs,
        owners: currentUser ? [currentUser.email] : [],
        description: '',
        rollbackPlan: ''
    };

    // Show loading state
    const submitBtn = document.getElementById('btn-quick-submit');
    submitBtn.disabled = true;
    document.querySelector('.btn-text').classList.add('hidden');
    document.querySelector('.btn-loading').classList.remove('hidden');

    let createdScheduleId = null;

    try {
        const token = localStorage.getItem('auth_token');

        // Task 2.2: Step 1 - Create schedule
        const scheduleResponse = await fetch(`${API_BASE_URL}/schedules`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(scheduleData)
        });

        // Task 2.3: Handle schedule creation failure
        if (!scheduleResponse.ok) {
            const error = await scheduleResponse.json();
            throw new Error(error.message || 'Failed to create schedule');
        }

        const schedule = await scheduleResponse.json();
        createdScheduleId = schedule.id;

        // Task 2.2: Step 2 - Assign to groups (if any selected)
        if (selectedGroupIds.length > 0) {
            try {
                const assignmentResponse = await fetch(`${API_BASE_URL}/schedules/${createdScheduleId}/groups`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`
                    },
                    body: JSON.stringify({
                        groupIds: selectedGroupIds,
                        assignedBy: currentUser ? currentUser.email : ''
                    })
                });

                // Task 2.4: Handle group assignment failure with rollback
                if (!assignmentResponse.ok) {
                    const assignError = await assignmentResponse.json();

                    // Attempt to delete the created schedule (rollback)
                    try {
                        const deleteResponse = await fetch(`${API_BASE_URL}/schedules/${createdScheduleId}`, {
                            method: 'DELETE',
                            headers: {
                                'Authorization': `Bearer ${token}`
                            }
                        });

                        if (!deleteResponse.ok) {
                            // Task 2.5: Rollback failure - orphaned schedule
                            console.error('Failed to rollback schedule creation after group assignment failure');
                            throw new Error(
                                `Schedule created but group assignment failed. Schedule ID: ${createdScheduleId}. ` +
                                `Please assign groups manually or delete the schedule. Error: ${assignError.message || 'Unknown error'}`
                            );
                        }

                        // Rollback successful
                        throw new Error(`Failed to assign groups: ${assignError.message || 'Unknown error'}. Schedule creation was rolled back.`);

                    } catch (rollbackError) {
                        // Re-throw the error (either rollback failure or assignment failure)
                        throw rollbackError;
                    }
                }
            } catch (groupError) {
                // Re-throw group assignment or rollback errors
                throw groupError;
            }
        }

        // Task 2.6: Success - close modal and refresh
        showToast('Schedule created successfully!', 'success');
        closeQuickCreate();

        // Refresh the schedule list
        if (typeof loadSchedules === 'function') {
            loadSchedules();
        }

    } catch (error) {
        console.error('Failed to create schedule:', error);

        // Task 2.3 & 2.4: Show error without closing modal
        showToast(error.message || 'Failed to create schedule', 'error');

        // Task 2.7: Preserve form values - don't reset form
        // Reset loading state only
        submitBtn.disabled = false;
        document.querySelector('.btn-text').classList.remove('hidden');
        document.querySelector('.btn-loading').classList.add('hidden');
    }
}

// ========================================
// Template Management
// ========================================

let templatesModal;
let templateFormModal;
let templateForm;
let currentTemplate = null;
let userTemplates = [];

// Initialize Template Management
document.addEventListener('DOMContentLoaded', () => {
    initializeTemplateManagement();
});

function initializeTemplateManagement() {
    templatesModal = document.getElementById('templates-modal');
    templateFormModal = document.getElementById('template-form-modal');
    templateForm = document.getElementById('template-form');
    
    // Button handlers
    document.getElementById('btn-manage-templates')?.addEventListener('click', openTemplatesModal);
    document.getElementById('templates-close')?.addEventListener('click', closeTemplatesModal);
    document.getElementById('btn-create-template')?.addEventListener('click', showCreateTemplateForm);
    document.getElementById('template-form-close')?.addEventListener('click', closeTemplateFormModal);
    document.getElementById('btn-template-cancel')?.addEventListener('click', closeTemplateFormModal);
    
    // Form submission
    templateForm?.addEventListener('submit', handleTemplateFormSubmit);
    
    // Click outside to close
    templatesModal?.addEventListener('click', (e) => {
        if (e.target === templatesModal) {
            closeTemplatesModal();
        }
    });
    
    templateFormModal?.addEventListener('click', (e) => {
        if (e.target === templateFormModal) {
            closeTemplateFormModal();
        }
    });
}

async function openTemplatesModal() {
    templatesModal.classList.remove('hidden');
    await loadTemplates();
}

function closeTemplatesModal() {
    templatesModal.classList.add('hidden');
}

function showCreateTemplateForm() {
    currentTemplate = null;
    document.getElementById('template-form-title').textContent = 'Create Template';
    resetTemplateForm();
    templateFormModal.classList.remove('hidden');
}

function showEditTemplateForm(template) {
    currentTemplate = template;
    document.getElementById('template-form-title').textContent = 'Edit Template';
    populateTemplateForm(template);
    templateFormModal.classList.remove('hidden');
}

function closeTemplateFormModal() {
    templateFormModal.classList.add('hidden');
    resetTemplateForm();
}

function resetTemplateForm() {
    templateForm.reset();
    document.getElementById('template-id').value = '';
    document.querySelectorAll('input[name="template-env"]').forEach(cb => cb.checked = false);
}

function populateTemplateForm(template) {
    document.getElementById('template-id').value = template.id;
    document.getElementById('template-name').value = template.name;
    document.getElementById('template-description').value = template.description || '';
    document.getElementById('template-service').value = template.serviceName || '';
    document.getElementById('template-rollback').value = template.rollbackPlan || '';
    document.getElementById('template-time-offset').value = template.defaultTimeOffset;
    
    // Set owners
    if (template.owners && template.owners.length > 0) {
        document.getElementById('template-owners').value = template.owners.join('; ');
    }
    
    // Set environments
    document.querySelectorAll('input[name="template-env"]').forEach(cb => {
        cb.checked = template.environments.includes(cb.value);
    });
}

async function loadTemplates() {
    try {
        const token = localStorage.getItem('auth_token');
        if (!token) {
            showToast('Please log in to view templates', 'error');
            return;
        }
        
        const response = await fetch(`${API_BASE_URL}/templates`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to load templates');
        }
        
        userTemplates = await response.json();
        renderTemplates();
    } catch (error) {
        console.error('Failed to load templates:', error);
        showToast('Failed to load templates', 'error');
    }
}

function renderTemplates() {
    const templatesList = document.getElementById('templates-list');
    const templatesEmpty = document.getElementById('templates-empty');
    
    if (!userTemplates || userTemplates.length === 0) {
        templatesList.innerHTML = '';
        templatesEmpty.classList.remove('hidden');
        return;
    }
    
    templatesEmpty.classList.add('hidden');
    
    templatesList.innerHTML = userTemplates.map(template => `
        <div class="template-card" data-template-id="${template.id}">
            <div class="template-header">
                <h3 class="template-name">${escapeHtml(template.name)}</h3>
                <div class="template-actions">
                    <button class="btn-icon btn-use" data-template-id="${template.id}" title="Use this template">
                        ⚡ Use
                    </button>
                    <button class="btn-icon btn-edit" data-template-id="${template.id}" title="Edit template">
                        ✏️
                    </button>
                    <button class="btn-icon btn-delete" data-template-id="${template.id}" title="Delete template">
                        🗑️
                    </button>
                </div>
            </div>
            ${template.description ? `<p class="template-description">${escapeHtml(template.description)}</p>` : ''}
            <div class="template-details">
                ${template.serviceName ? `<div class="template-detail">
                    <span class="detail-label">Service:</span>
                    <span class="detail-value">${escapeHtml(template.serviceName)}</span>
                </div>` : ''}
                <div class="template-detail">
                    <span class="detail-label">Environments:</span>
                    <span class="detail-value">
                        ${template.environments.map(env => 
                            `<span class="badge badge-env-${env}">${env}</span>`
                        ).join(' ')}
                    </span>
                </div>
                <div class="template-detail">
                    <span class="detail-label">Time:</span>
                    <span class="detail-value">${formatTimeOffset(template.defaultTimeOffset)}</span>
                </div>
                ${template.owners && template.owners.length > 0 ? `<div class="template-detail">
                    <span class="detail-label">Owners:</span>
                    <span class="detail-value">${template.owners.length} owner(s)</span>
                </div>` : ''}
            </div>
        </div>
    `).join('');
    
    // Attach event listeners
    templatesList.querySelectorAll('.btn-use').forEach(btn => {
        btn.addEventListener('click', () => {
            const templateId = btn.dataset.templateId;
            useTemplate(templateId);
        });
    });
    
    templatesList.querySelectorAll('.btn-edit').forEach(btn => {
        btn.addEventListener('click', () => {
            const templateId = btn.dataset.templateId;
            const template = userTemplates.find(t => t.id === templateId);
            if (template) showEditTemplateForm(template);
        });
    });
    
    templatesList.querySelectorAll('.btn-delete').forEach(btn => {
        btn.addEventListener('click', () => {
            const templateId = btn.dataset.templateId;
            deleteTemplate(templateId);
        });
    });
}

function formatTimeOffset(minutes) {
    if (minutes === 0) return 'Now';
    if (minutes < 60) return `+${minutes} minutes`;
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    if (remainingMinutes === 0) return `+${hours} hour${hours > 1 ? 's' : ''}`;
    return `+${hours}h ${remainingMinutes}m`;
}

async function handleTemplateFormSubmit(e) {
    e.preventDefault();
    
    const templateId = document.getElementById('template-id').value;
    const name = document.getElementById('template-name').value.trim();
    const description = document.getElementById('template-description').value.trim();
    const serviceName = document.getElementById('template-service').value.trim();
    const rollbackPlan = document.getElementById('template-rollback').value.trim();
    const defaultTimeOffset = parseInt(document.getElementById('template-time-offset').value);
    
    // Get selected environments
    const environments = Array.from(document.querySelectorAll('input[name="template-env"]:checked'))
        .map(cb => cb.value);
    
    // Get owners
    const ownersInput = document.getElementById('template-owners').value.trim();
    const owners = ownersInput ? ownersInput.split(';').map(o => o.trim()).filter(o => o) : [];
    
    // Validation
    if (!name) {
        showToast('Template name is required', 'error');
        return;
    }
    
    if (environments.length === 0) {
        showToast('At least one environment is required', 'error');
        return;
    }
    
    const templateData = {
        name,
        description: description || undefined,
        serviceName: serviceName || undefined,
        environments,
        owners: owners.length > 0 ? owners : undefined,
        rollbackPlan: rollbackPlan || undefined,
        defaultTimeOffset
    };
    
    try {
        const token = localStorage.getItem('auth_token');
        const url = templateId 
            ? `${API_BASE_URL}/templates/${templateId}`
            : `${API_BASE_URL}/templates`;
        const method = templateId ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(templateData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to save template');
        }
        
        showToast(`Template ${templateId ? 'updated' : 'created'} successfully!`, 'success');
        closeTemplateFormModal();
        await loadTemplates();
    } catch (error) {
        console.error('Failed to save template:', error);
        showToast(error.message || 'Failed to save template', 'error');
    }
}

async function deleteTemplate(templateId) {
    const template = userTemplates.find(t => t.id === templateId);
    if (!template) return;
    
    if (!confirm(`Are you sure you want to delete the template "${template.name}"?`)) {
        return;
    }
    
    try {
        const token = localStorage.getItem('auth_token');
        const response = await fetch(`${API_BASE_URL}/templates/${templateId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to delete template');
        }
        
        showToast('Template deleted successfully', 'success');
        await loadTemplates();
    } catch (error) {
        console.error('Failed to delete template:', error);
        showToast('Failed to delete template', 'error');
    }
}

function useTemplate(templateId) {
    const template = userTemplates.find(t => t.id === templateId);
    if (!template) return;
    
    // Close templates modal
    closeTemplatesModal();
    
    // Open Quick Create with template data
    openQuickCreateWithTemplate(template);
}

function openQuickCreateWithTemplate(template) {
    // Open Quick Create modal
    quickCreateModal.classList.remove('hidden');
    
    // Pre-fill form with template data
    if (template.serviceName) {
        document.getElementById('quick-service').value = template.serviceName;
    }
    
    // Set environments
    document.querySelectorAll('input[name="quick-env"]').forEach(cb => {
        cb.checked = template.environments.includes(cb.value);
    });
    
    // Set time offset
    const timeOffset = template.defaultTimeOffset;
    selectedTimeOffset = timeOffset;
    
    // Update time button states
    document.querySelectorAll('.time-btn').forEach(btn => {
        btn.classList.remove('active');
        if (btn.dataset.offset === String(timeOffset)) {
            btn.classList.add('active');
        }
    });
    
    // Hide custom time input
    document.getElementById('quick-time-custom').classList.add('hidden');
    
    // Focus on service input
    setTimeout(() => {
        document.getElementById('quick-service')?.focus();
    }, 100);
    
    showToast(`Using template: ${template.name}`, 'success');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Initialize tag inputs after classes are loaded
document.addEventListener('DOMContentLoaded', () => {
    initializeFormTagInputs();
});
