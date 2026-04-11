// API Configuration
const API_BASE_URL = 'http://localhost:8080';

// Application State
let currentSchedule = null;
let isEditMode = false;
let allGroups = [];
let selectedGroupId = 'all'; // 'all', 'ungrouped', or group UUID
let dateSectionCollapseStates = {}; // Stores collapse state for date sections

// Authenticated user profile, populated from GET /users/me on bootstrap.
// This is the single source of truth for user identity; do not read identity
// from the filter-owner input or from localStorage.
let currentUser = null;

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
document.addEventListener('DOMContentLoaded', async () => {
    setupEventListeners();
    initializeFormTagInputs();

    // Initialize controllers (GroupController is defined later in the file,
    // but classes/functions are hoisted and available by the time this fires).
    groupController = new GroupController();
    window.groupController = groupController;

    loadDateSectionCollapseStates();
    loadSelectedGroupFromURL();

    // Resolve the authenticated user profile before loading any data that
    // depends on identity (groups visible to user, schedules owned by user).
    await fetchCurrentUser();

    // Render the user chip (renderUserChip is defined later in the file).
    if (typeof renderUserChip === 'function') {
        renderUserChip(currentUser);
    }

    loadGroupsAndRenderSidebar();
    loadSchedules();
});

// Event Listeners
function setupEventListeners() {
    // Toolbar buttons
    document.getElementById('btn-create').addEventListener('click', showCreateForm);
    document.getElementById('btn-refresh').addEventListener('click', () => loadSchedules());
    document.getElementById('btn-cancel').addEventListener('click', showListView);
    document.getElementById('btn-back').addEventListener('click', showListView);

    // Sidebar controls
    document.getElementById('btn-create-group-sidebar')?.addEventListener('click', () => {
        if (groupController) {
            groupController.showCreateGroupModal();
        }
    });

    document.querySelector('.sidebar-item--all')?.addEventListener('click', () => selectGroup('all'));
    document.querySelector('.sidebar-item--ungrouped')?.addEventListener('click', () => selectGroup('ungrouped'));

    // Mobile sidebar controls
    document.getElementById('hamburger-menu')?.addEventListener('click', openMobileSidebar);
    document.getElementById('sidebar-backdrop')?.addEventListener('click', closeMobileSidebar);

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

// ===================================
// AUTHENTICATED USER BOOTSTRAP
// ===================================

// Fetch the authenticated user profile from the backend and populate the
// module-level `currentUser` state. Called once on app bootstrap before any
// data loads. Returns the user object on success, or null if the profile
// could not be resolved (in which case the chip renders a minimal variant).
//
// Error policy (per design.md Decision 2):
//   401 -> token invalid; clear local state and redirect to sign-in.
//   404 -> endpoint missing; warn and fall back to minimal profile.
//   5xx or network error -> notify and fall back to minimal profile.
async function fetchCurrentUser() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/auth/google/login';
        return null;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/users/me`, {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
        });

        if (response.status === 401) {
            localStorage.removeItem('auth_token');
            localStorage.removeItem('user_email');
            localStorage.removeItem('user_name');
            localStorage.removeItem('user_role');
            window.location.href = '/auth/google/login';
            return null;
        }

        if (response.status === 404) {
            console.warn('fetchCurrentUser: /users/me returned 404; falling back to minimal profile');
            currentUser = { email: null, name: null, role: null, _minimal: true };
            return currentUser;
        }

        if (!response.ok) {
            console.error(`fetchCurrentUser: unexpected status ${response.status}`);
            if (typeof showNotification === 'function') {
                showNotification('Could not load your profile. Some features may be limited.', 'error');
            }
            currentUser = { email: null, name: null, role: null, _minimal: true };
            return currentUser;
        }

        currentUser = await response.json();
        return currentUser;
    } catch (err) {
        console.error('fetchCurrentUser failed:', err);
        if (typeof showNotification === 'function') {
            showNotification('Could not load your profile. Some features may be limited.', 'error');
        }
        currentUser = { email: null, name: null, role: null, _minimal: true };
        return currentUser;
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
        const ownerEmail = getCurrentUser();
        if (!ownerEmail) {
            showNotification('You must be signed in to assign groups', 'error');
            return;
        }
        const [schedule, groupsResponse] = await Promise.all([
            getSchedule(scheduleId),
            fetch(`${API_BASE_URL}/groups?owner=${encodeURIComponent(ownerEmail)}`, {
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
        const overlayEl = overlay.firstElementChild;
        document.body.appendChild(overlayEl);

        // Install focus trap on the quick-assign overlay (task 11.4).
        _quickAssignRelease = trapFocus(overlayEl, () => closeQuickAssign());

    } catch (error) {
        console.error('Quick assign error:', error);
        showNotification('Failed to load groups: ' + error.message, 'error');
    }
}

// Holds the focus-trap release function for the dynamically-created
// quick-assign overlay. Set when the overlay is shown, called when closed.
let _quickAssignRelease = null;

function closeQuickAssign(event) {
    if (event && event.target.id !== 'quick-assign-overlay' && !event.target.classList.contains('modal-close')) {
        return;
    }
    const overlay = document.getElementById('quick-assign-overlay');
    if (overlay) {
        overlay.remove();
    }
    if (_quickAssignRelease) {
        _quickAssignRelease();
        _quickAssignRelease = null;
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

    // Get selected group IDs
    const selectedGroupIds = Array.from(document.querySelectorAll('input[name="form-group"]:checked'))
        .map(cb => cb.value);

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

    let createdScheduleId = null;

    try {
        let schedule;
        if (isEditMode) {
            schedule = await updateSchedule(currentSchedule.id, data);

            // Update group assignments for edited schedule
            if (selectedGroupIds.length > 0) {
                try {
                    const token = localStorage.getItem('auth_token');
                    await fetch(`${API_BASE_URL}/schedules/${currentSchedule.id}/groups`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': `Bearer ${token}`
                        },
                        body: JSON.stringify({ groupIds: selectedGroupIds })
                    });
                } catch (groupError) {
                    console.error('Failed to update group assignments:', groupError);
                    showNotification('Schedule updated, but group assignment failed', 'error');
                }
            }
        } else {
            // Create new schedule
            schedule = await createSchedule(data);
            createdScheduleId = schedule.id;

            // Step 2 - Assign to groups (if any selected)
            if (selectedGroupIds.length > 0) {
                try {
                    const token = localStorage.getItem('auth_token');
                    const assignmentResponse = await fetch(`${API_BASE_URL}/schedules/${createdScheduleId}/groups`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': `Bearer ${token}`
                        },
                        body: JSON.stringify({ groupIds: selectedGroupIds })
                    });

                    // Handle group assignment failure with rollback
                    if (!assignmentResponse.ok) {
                        const assignError = await assignmentResponse.json();

                        // Attempt to delete the created schedule (rollback)
                        try {
                            await fetch(`${API_BASE_URL}/schedules/${createdScheduleId}`, {
                                method: 'DELETE',
                                headers: {
                                    'Authorization': `Bearer ${token}`
                                }
                            });

                            throw new Error(`Failed to assign groups: ${assignError.message || 'Unknown error'}. Schedule creation was rolled back.`);
                        } catch (rollbackError) {
                            // Rollback failed - orphaned schedule
                            throw new Error(
                                `Schedule created but group assignment failed. Schedule ID: ${createdScheduleId}. ` +
                                `Please assign groups manually or delete the schedule. Error: ${assignError.message || 'Unknown error'}`
                            );
                        }
                    }
                } catch (groupError) {
                    throw groupError;
                }
            }
        }

        showDetail(schedule);
    } catch (error) {
        // Error already shown in API call or thrown above
        if (error.message) {
            showNotification(error.message, 'error');
        }
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
// Keep track of the active notification timeout so repeated calls don't leave
// stale text sitting in the live region.
let _notificationTimeoutId = null;

function showNotification(message, type = 'success') {
    // Route errors to the assertive live region so screen readers announce them
    // immediately. Route success/info/default to the polite region so they
    // queue without interrupting the user's current activity.
    const assertiveRegion = document.getElementById('notification-assertive');

    if (type === 'error' && assertiveRegion) {
        // Clear first so a repeat of the same message still triggers an announcement.
        assertiveRegion.textContent = '';
        // Let the DOM settle before writing the new message.
        requestAnimationFrame(() => {
            assertiveRegion.textContent = message;
        });
    }

    // Visual notification (visible to sighted users) always uses the primary element.
    notification.textContent = message;
    notification.className = `notification ${type}`;
    notification.classList.remove('hidden');

    if (_notificationTimeoutId) {
        clearTimeout(_notificationTimeoutId);
    }
    _notificationTimeoutId = setTimeout(() => {
        notification.classList.add('hidden');
        // Clear text content so assistive tech does not re-announce the
        // stale message when the region becomes relevant again.
        notification.textContent = '';
        if (assertiveRegion) {
            assertiveRegion.textContent = '';
        }
        _notificationTimeoutId = null;
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
// SIDEBAR NAVIGATION FUNCTIONALITY
// ===================================

// Load groups and render sidebar
async function loadGroupsAndRenderSidebar() {
    try {
        const token = localStorage.getItem('auth_token');
        if (!token) {
            allGroups = [];
            renderSidebar();
            return;
        }

        const owner = encodeURIComponent(getCurrentUser());
        const response = await fetch(`${API_BASE_URL}/groups?owner=${owner}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            allGroups = await response.json();
            // Sort: favorites first, then alphabetically
            allGroups.sort((a, b) => {
                if (a.isFavorite && !b.isFavorite) return -1;
                if (!a.isFavorite && b.isFavorite) return 1;
                return a.name.localeCompare(b.name);
            });
        } else {
            allGroups = [];
        }
    } catch (error) {
        console.error('Failed to load groups:', error);
        allGroups = [];
    } finally {
        renderSidebar();
    }
}

// Render sidebar with groups
function renderSidebar() {
    const sidebarList = document.getElementById('sidebar-groups-list');
    if (!sidebarList) return;

    sidebarList.innerHTML = '';

    if (allGroups.length === 0) {
        sidebarList.innerHTML = '<p style="padding: 16px; color: var(--color-text-secondary); font-size: 14px; text-align: center;">No groups yet.<br>Create your first group!</p>';
        updateSidebarSelection();
        return;
    }

    allGroups.forEach(group => {
        const item = document.createElement('button');
        item.className = 'sidebar-item';
        item.dataset.groupId = group.id;

        const visibilityIcon = group.visibility === 'public' ? '🌐' : '🔒';
        const visibilityClass = group.visibility === 'public' ? 'sidebar-item__icon--public' : 'sidebar-item__icon--private';
        const favoriteIcon = group.isFavorite ? '★' : '';

        item.innerHTML = `
            <span class="sidebar-item__name">
                ${favoriteIcon ? `<span class="sidebar-item__favorite">${favoriteIcon}</span>` : ''}
                <span class="sidebar-item__icon ${visibilityClass}">${visibilityIcon}</span>
                ${escapeHtml(group.name)}
            </span>
            <span class="sidebar-item__count">0</span>
            <button class="sidebar-item__settings" title="Settings" onclick="event.stopPropagation(); editGroupFromSidebar('${group.id}')">
                ⚙️
            </button>
        `;

        item.addEventListener('click', () => selectGroup(group.id));

        sidebarList.appendChild(item);
    });

    updateSidebarSelection();
}

// Update sidebar selection highlighting
function updateSidebarSelection() {
    document.querySelectorAll('.sidebar-item').forEach(item => {
        const groupId = item.dataset.groupId;
        if (groupId === selectedGroupId) {
            item.classList.add('sidebar-item--active');
        } else {
            item.classList.remove('sidebar-item--active');
        }
    });
}

// Select a group (or pseudo-group)
function selectGroup(groupId) {
    selectedGroupId = groupId;
    updateURLForGroup(groupId);
    updateSidebarSelection();
    closeMobileSidebar();
    loadSchedules();
}

// Load selected group from URL hash
function loadSelectedGroupFromURL() {
    const hash = window.location.hash;
    if (hash.startsWith('#group/')) {
        selectedGroupId = hash.substring(7); // Remove '#group/'
    } else if (hash === '#all') {
        selectedGroupId = 'all';
    } else if (hash === '#ungrouped') {
        selectedGroupId = 'ungrouped';
    } else {
        selectedGroupId = 'all';
        window.location.hash = '#all';
    }
}

// Update URL hash when group is selected
function updateURLForGroup(groupId) {
    if (groupId === 'all') {
        window.location.hash = '#all';
    } else if (groupId === 'ungrouped') {
        window.location.hash = '#ungrouped';
    } else {
        window.location.hash = `#group/${groupId}`;
    }
}

// Handle browser back/forward
window.addEventListener('hashchange', () => {
    loadSelectedGroupFromURL();
    updateSidebarSelection();
    loadSchedules();
});

// Mobile sidebar controls
function openMobileSidebar() {
    const sidebar = document.getElementById('groups-sidebar');
    const backdrop = document.getElementById('sidebar-backdrop');
    sidebar.classList.add('sidebar--open');
    backdrop.classList.add('sidebar-backdrop--visible');
}

function closeMobileSidebar() {
    const sidebar = document.getElementById('groups-sidebar');
    const backdrop = document.getElementById('sidebar-backdrop');
    sidebar.classList.remove('sidebar--open');
    backdrop.classList.remove('sidebar-backdrop--visible');
}

// Edit group from sidebar
function editGroupFromSidebar(groupId) {
    const group = allGroups.find(g => g.id === groupId);
    if (group && groupController) {
        groupController.showEditGroupModal(group);
    }
}

// Update sidebar counts based on schedules
function updateSidebarCounts(schedules) {
    // Count all schedules
    const allCount = schedules.length;
    document.querySelector('.sidebar-item--all .sidebar-item__count').textContent = allCount;

    // Count ungrouped
    const ungroupedCount = schedules.filter(s => !s.groups || s.groups.length === 0).length;
    document.querySelector('.sidebar-item--ungrouped .sidebar-item__count').textContent = ungroupedCount;

    // Count per group
    allGroups.forEach(group => {
        const count = schedules.filter(s =>
            s.groups && s.groups.some(g => g.id === group.id)
        ).length;
        const item = document.querySelector(`.sidebar-item[data-group-id="${group.id}"]`);
        if (item) {
            item.querySelector('.sidebar-item__count').textContent = count;
        }
    });
}

// ===================================
// DATE GROUPING FUNCTIONALITY
// ===================================

// Group schedules by relative date
function groupSchedulesByDate(schedules) {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);
    const nextWeek = new Date(today);
    nextWeek.setDate(nextWeek.getDate() + 7);

    const groups = {
        today: [],
        tomorrow: [],
        thisWeek: [],
        later: []
    };

    schedules.forEach(schedule => {
        const scheduledDate = new Date(schedule.scheduledAt);
        const scheduleDay = new Date(scheduledDate.getFullYear(), scheduledDate.getMonth(), scheduledDate.getDate());

        if (scheduleDay.getTime() === today.getTime()) {
            groups.today.push(schedule);
        } else if (scheduleDay.getTime() === tomorrow.getTime()) {
            groups.tomorrow.push(schedule);
        } else if (scheduleDay < nextWeek) {
            groups.thisWeek.push(schedule);
        } else {
            groups.later.push(schedule);
        }
    });

    // Sort schedules within each group by time
    Object.keys(groups).forEach(key => {
        groups[key].sort((a, b) => new Date(a.scheduledAt) - new Date(b.scheduledAt));
    });

    return groups;
}

// Render date-grouped schedules
function renderDateGroupedSchedules(schedules) {
    const container = document.getElementById('date-grouped-schedules');
    if (!container) return;

    container.innerHTML = '';

    if (!schedules || schedules.length === 0) {
        container.innerHTML = '<p style="text-align: center; color: var(--color-text-secondary); padding: 40px;">No schedules found</p>';
        return;
    }

    const dateGroups = groupSchedulesByDate(schedules);
    const sections = [
        { key: 'today', label: 'Today' },
        { key: 'tomorrow', label: 'Tomorrow' },
        { key: 'thisWeek', label: 'This Week' },
        { key: 'later', label: 'Later' }
    ];

    sections.forEach(section => {
        const schedules = dateGroups[section.key];
        if (schedules.length > 0) {
            const sectionEl = renderDateSection(section.key, section.label, schedules);
            container.appendChild(sectionEl);
        }
    });
}

// Render a single date section
function renderDateSection(sectionKey, sectionLabel, schedules) {
    const isCollapsed = dateSectionCollapseStates[sectionKey] || false;

    const section = document.createElement('div');
    section.className = `date-section ${isCollapsed ? 'date-section--collapsed' : ''}`;
    section.dataset.sectionKey = sectionKey;

    const header = document.createElement('div');
    header.className = 'date-section-header';
    header.innerHTML = `
        <div class="date-section-header__title">
            <span>${sectionLabel}</span>
            <span class="date-section-header__count">${schedules.length}</span>
        </div>
        <span class="date-section-header__chevron">▼</span>
    `;
    header.addEventListener('click', () => toggleDateSection(sectionKey));

    const body = document.createElement('div');
    body.className = 'date-section-body';

    const schedulesList = document.createElement('div');
    schedulesList.className = 'date-section-schedules';

    schedules.forEach(schedule => {
        const item = renderDateScheduleItem(schedule);
        schedulesList.appendChild(item);
    });

    body.appendChild(schedulesList);
    section.appendChild(header);
    section.appendChild(body);

    return section;
}

// Render a single schedule item in date section
function renderDateScheduleItem(schedule) {
    const item = document.createElement('div');
    item.className = 'date-schedule-item';
    item.onclick = () => loadAndShowDetail(schedule.id);

    const scheduledTime = new Date(schedule.scheduledAt);
    const now = new Date();
    const isOverdue = scheduledTime < now && schedule.status === 'created';

    const timeStr = formatTime(scheduledTime);
    const timeClass = isOverdue ? 'date-schedule-item__time--overdue' : '';

    item.innerHTML = `
        <div class="date-schedule-item__time ${timeClass}">
            ${timeStr}
            ${isOverdue ? '<div class="overdue-badge">OVERDUE</div>' : ''}
        </div>
        <div class="date-schedule-item__info">
            <div class="date-schedule-item__service">${escapeHtml(schedule.serviceName)}</div>
            <div class="date-schedule-item__meta">
                <div class="date-schedule-item__owners">${getOwnersDisplay(schedule.owners, 2)}</div>
                ${getEnvironmentBadges(schedule.environments)}
            </div>
        </div>
        <div class="date-schedule-item__status">
            ${getStatusIcon(schedule.status)}${getStatusBadge(schedule.status)}
        </div>
    `;

    return item;
}

// Toggle date section collapse
function toggleDateSection(sectionKey) {
    const section = document.querySelector(`.date-section[data-section-key="${sectionKey}"]`);
    if (!section) return;

    const isCollapsed = section.classList.toggle('date-section--collapsed');
    dateSectionCollapseStates[sectionKey] = isCollapsed;
    saveDateSectionCollapseStates();
}

// Save collapse states to localStorage
function saveDateSectionCollapseStates() {
    try {
        localStorage.setItem('dateSectionCollapseStates', JSON.stringify(dateSectionCollapseStates));
    } catch (e) {
        console.error('Failed to save collapse states:', e);
    }
}

// Load collapse states from localStorage
function loadDateSectionCollapseStates() {
    try {
        const saved = localStorage.getItem('dateSectionCollapseStates');
        if (saved) {
            dateSectionCollapseStates = JSON.parse(saved);
        }
    } catch (e) {
        console.error('Failed to load collapse states:', e);
        dateSectionCollapseStates = {};
    }
}

// Format time as HH:MM
function formatTime(date) {
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${hours}:${minutes}`;
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

async function createGroupWithVisibility(name, description, owner, visibility) {
    return await apiCall('/groups', {
        method: 'POST',
        body: JSON.stringify({ name, description, owner, visibility })
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

async function updateGroupWithVisibility(id, name, description, visibility) {
    return await apiCall(`/groups/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ name, description, visibility })
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
        // Focus-trap release functions, set when a modal is shown and called
        // when the modal is hidden (task 11.2, 11.3).
        this._groupModalRelease = null;
        this._assignGroupModalRelease = null;
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
        document.getElementById('group-visibility-private').checked = true;
        this.groupModal.classList.remove('hidden');
        this._groupModalRelease = trapFocus(this.groupModal, () => this.hideGroupModal());
    }

    async showEditGroupModal(group) {
        this.currentGroupId = group.id;
        document.getElementById('group-modal-title').textContent = 'Edit Group';
        document.getElementById('group-name').value = group.name;
        document.getElementById('group-description').value = group.description || '';
        document.getElementById('group-id').value = group.id;

        // Set visibility
        if (group.visibility === 'public') {
            document.getElementById('group-visibility-public').checked = true;
        } else {
            document.getElementById('group-visibility-private').checked = true;
        }

        this.groupModal.classList.remove('hidden');
        this._groupModalRelease = trapFocus(this.groupModal, () => this.hideGroupModal());
    }

    hideGroupModal() {
        this.groupModal.classList.add('hidden');
        this.groupForm.reset();
        if (this._groupModalRelease) {
            this._groupModalRelease();
            this._groupModalRelease = null;
        }
    }

    async showGroupsListModal() {
        try {
            // Always source identity from getCurrentUser(); never from the
            // filter-owner input (which is a search filter, not identity).
            const owner = getCurrentUser();
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
            this._assignGroupModalRelease = trapFocus(this.assignGroupModal, () => this.hideAssignGroupModal());
        } catch (error) {
            showNotification('Failed to load groups: ' + error.message, 'error');
        }
    }

    hideAssignGroupModal() {
        this.assignGroupModal.classList.add('hidden');
        if (this._assignGroupModalRelease) {
            this._assignGroupModalRelease();
            this._assignGroupModalRelease = null;
        }
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
        const visibility = document.querySelector('input[name="group-visibility"]:checked').value;
        const owner = getCurrentUser();

        try {
            if (this.currentGroupId) {
                await updateGroupWithVisibility(this.currentGroupId, name, description, visibility);
                showNotification('Group updated successfully', 'success');
            } else {
                await createGroupWithVisibility(name, description, owner, visibility);
                showNotification('Group created successfully', 'success');
            }

            this.hideGroupModal();

            // Refresh sidebar
            await loadGroupsAndRenderSidebar();

            if (this.groupsListModal && !this.groupsListModal.classList.contains('hidden')) {
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

            // Refresh sidebar
            await loadGroupsAndRenderSidebar();

            // If we were viewing this group, switch to "All"
            if (selectedGroupId === id) {
                selectGroup('all');
            }

            if (this.groupsListModal && !this.groupsListModal.classList.contains('hidden')) {
                await this.showGroupsListModal();
            }
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

            // Refresh sidebar
            await loadGroupsAndRenderSidebar();

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
// UTILITY FUNCTIONS
// ===================================

// Returns the authenticated user's identifier (email) for use in API calls
// that need ownership attribution. Sources identity from the module-level
// `currentUser` state, which is populated by fetchCurrentUser() on bootstrap.
// NEVER read identity from the `filter-owner` input field -- that field is a
// search filter, not an identity source.
function getCurrentUser() {
    return currentUser?.email ?? null;
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

// Update loadSchedules to work with sidebar groups and date grouping
const originalLoadSchedules = loadSchedules;
loadSchedules = async function() {
    const owner = document.getElementById('filter-owner').value;
    const status = document.getElementById('filter-status').value;
    const environment = document.getElementById('filter-env').value;

    const params = new URLSearchParams();
    if (owner) params.append('owner', owner);
    if (status) params.append('status', status);
    if (environment) params.append('environment', environment);

    const query = params.toString() ? `?${params.toString()}` : '';

    try {
        let schedules = await apiCall(`/schedules${query}`);

        // Filter by selected group
        if (selectedGroupId === 'ungrouped') {
            schedules = schedules.filter(s => !s.groups || s.groups.length === 0);
        } else if (selectedGroupId !== 'all') {
            schedules = schedules.filter(s =>
                s.groups && s.groups.some(g => g.id === selectedGroupId)
            );
        }

        // Update sidebar counts with all schedules (before filtering)
        updateSidebarCounts(await apiCall(`/schedules${query}`));

        // Render using date grouping
        renderDateGroupedSchedules(schedules);
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
// Group Selector for Standard Form
// ========================================

let formGroups = [];
let formGroupsLoading = false;

// Load groups for standard form
async function loadGroupsForForm() {
    const loadingEl = document.getElementById('form-groups-loading');
    const emptyEl = document.getElementById('form-groups-empty');
    const listEl = document.getElementById('form-groups-list');

    // Show loading state
    loadingEl.classList.remove('hidden');
    emptyEl.classList.add('hidden');
    listEl.innerHTML = '';
    formGroupsLoading = true;

    try {
        const token = localStorage.getItem('auth_token');
        if (!token) {
            formGroups = [];
            renderFormGroupSelector();
            return;
        }

        const owner = encodeURIComponent(getCurrentUser());
        const response = await fetch(`${API_BASE_URL}/groups?owner=${owner}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            formGroups = await response.json();
            // Groups are already ordered with favorites first by the API
        } else {
            formGroups = [];
        }
    } catch (error) {
        console.error('Failed to load groups for form:', error);
        formGroups = [];
    } finally {
        formGroupsLoading = false;
        loadingEl.classList.add('hidden');
        renderFormGroupSelector();
    }
}

// Render group checkboxes in standard form
function renderFormGroupSelector() {
    const emptyEl = document.getElementById('form-groups-empty');
    const listEl = document.getElementById('form-groups-list');

    listEl.innerHTML = '';

    // Empty state
    if (!formGroups || formGroups.length === 0) {
        emptyEl.classList.remove('hidden');
        return;
    }

    emptyEl.classList.add('hidden');

    // Render checkboxes with favorite indicators
    formGroups.forEach((group, index) => {
        const div = document.createElement('div');
        div.className = 'checkbox-item';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.id = `form-group-${group.id}`;
        checkbox.value = group.id;
        checkbox.name = 'form-group';

        const label = document.createElement('label');
        label.htmlFor = `form-group-${group.id}`;

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

        // Keyboard navigation support
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

// Update showCreateForm to load groups
const originalShowCreateForm = showCreateForm;
showCreateForm = function() {
    originalShowCreateForm();
    loadGroupsForForm();
};

// Update showEditForm to load groups and pre-select assigned groups
const originalShowEditForm = showEditForm;
showEditForm = async function() {
    originalShowEditForm();
    await loadGroupsForForm();
    
    // Pre-select groups that the schedule is already assigned to
    if (currentSchedule && currentSchedule.id) {
        try {
            const token = localStorage.getItem('auth_token');
            const response = await fetch(`${API_BASE_URL}/schedules/${currentSchedule.id}/groups`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (response.ok) {
                const assignedGroups = await response.json();
                const assignedGroupIds = assignedGroups.map(g => g.id);

                // Check the boxes for assigned groups
                document.querySelectorAll('input[name="form-group"]').forEach(cb => {
                    cb.checked = assignedGroupIds.includes(cb.value);
                });
            }
        } catch (error) {
            console.error('Failed to load assigned groups:', error);
        }
    }
};

// ===================================
// USER CHIP COMPONENT
// ===================================

// Returns the initials (up to 2 characters) derived from a display name or
// email. Used as the fallback avatar for the chip.
function _chipInitials(user) {
    const source = (user?.name || user?.email || '').trim();
    if (!source) return '?';
    const parts = source.split(/[\s@._-]+/).filter(Boolean);
    if (parts.length === 0) return source.charAt(0).toUpperCase();
    if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
    return (parts[0].charAt(0) + parts[1].charAt(0)).toUpperCase();
}

// Returns a plain-language description of a role for the title tooltip.
function _roleDescription(role) {
    switch (role) {
        case 'admin':    return 'Administrator: full access including approvals and user management';
        case 'deployer': return 'Deployer: can create, update, and manage own schedules';
        case 'viewer':   return 'Viewer: read-only access to schedules';
        default:         return 'User role';
    }
}

// Render the user chip into the `#user-chip-container` element.
// Called once on bootstrap after fetchCurrentUser() resolves.
//
// If `user` is null or has `_minimal: true`, a minimal chip is rendered that
// shows only a sign-out control (task 6.7).
function renderUserChip(user) {
    const container = document.getElementById('user-chip-container');
    if (!container) return;

    // Minimal variant: profile could not be loaded. Show only a sign-out control.
    if (!user || user._minimal || (!user.email && !user.name)) {
        container.innerHTML = `
            <button type="button" class="user-chip user-chip--minimal" id="user-chip-signout-only" aria-label="Sign out">
                <span class="user-chip__initials" aria-hidden="true">?</span>
                <span class="user-chip__signout-label">Sign out</span>
            </button>
        `;
        document.getElementById('user-chip-signout-only').addEventListener('click', signOut);
        return;
    }

    const displayName = user.name || user.email || 'User';
    const role = (user.role || 'viewer').toLowerCase();
    const initials = _chipInitials(user);
    const roleDesc = _roleDescription(role);

    container.innerHTML = `
        <button type="button"
                class="user-chip"
                id="user-chip-button"
                aria-haspopup="menu"
                aria-expanded="false"
                aria-controls="user-chip-menu"
                aria-label="${escapeHtml(displayName)}, role: ${escapeHtml(role)}, user menu">
            <span class="user-chip__avatar" aria-hidden="true">${escapeHtml(initials)}</span>
            <span class="user-chip__body">
                <span class="user-chip__name" title="${escapeHtml(displayName)}">${escapeHtml(displayName)}</span>
                <span class="user-chip__role user-chip__role--${escapeHtml(role)}" title="${escapeHtml(roleDesc)}">${escapeHtml(role)}</span>
            </span>
            <span class="user-chip__caret" aria-hidden="true">▾</span>
        </button>
        <div class="user-chip__menu hidden"
             id="user-chip-menu"
             role="menu"
             aria-labelledby="user-chip-button">
            <div class="user-chip__menu-info" role="presentation">
                <div class="user-chip__menu-email">${escapeHtml(user.email || '')}</div>
                <div class="user-chip__menu-role" title="${escapeHtml(roleDesc)}">Role: ${escapeHtml(role)}</div>
            </div>
            <div class="user-chip__menu-separator" role="separator"></div>
            <button type="button" class="user-chip__menu-item" role="menuitem" id="user-chip-signout">Sign out</button>
        </div>
    `;

    const chipButton = document.getElementById('user-chip-button');
    const chipMenu = document.getElementById('user-chip-menu');
    const signOutBtn = document.getElementById('user-chip-signout');

    chipButton.addEventListener('click', (e) => {
        e.stopPropagation();
        toggleUserChipMenu();
    });

    chipButton.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
            e.preventDefault();
            openUserChipMenu();
            signOutBtn?.focus();
        }
    });

    signOutBtn.addEventListener('click', signOut);

    // Arrow-key navigation within the menu. Currently only one menu item
    // (Sign out), but the scaffolding is in place for future items.
    chipMenu.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            e.preventDefault();
            closeUserChipMenu();
            chipButton.focus();
            return;
        }
        if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
            e.preventDefault();
            const items = chipMenu.querySelectorAll('[role="menuitem"]');
            if (items.length === 0) return;
            const currentIndex = Array.from(items).indexOf(document.activeElement);
            let nextIndex;
            if (e.key === 'ArrowDown') {
                nextIndex = (currentIndex + 1) % items.length;
            } else {
                nextIndex = (currentIndex - 1 + items.length) % items.length;
            }
            items[nextIndex].focus();
        }
    });

    // Click-outside closes the menu.
    document.addEventListener('click', (e) => {
        if (!chipMenu.classList.contains('hidden') &&
            !chipButton.contains(e.target) &&
            !chipMenu.contains(e.target)) {
            closeUserChipMenu();
        }
    });
}

function toggleUserChipMenu() {
    const menu = document.getElementById('user-chip-menu');
    if (!menu) return;
    if (menu.classList.contains('hidden')) {
        openUserChipMenu();
    } else {
        closeUserChipMenu();
    }
}

function openUserChipMenu() {
    const menu = document.getElementById('user-chip-menu');
    const button = document.getElementById('user-chip-button');
    if (!menu || !button) return;
    menu.classList.remove('hidden');
    button.setAttribute('aria-expanded', 'true');
}

function closeUserChipMenu() {
    const menu = document.getElementById('user-chip-menu');
    const button = document.getElementById('user-chip-button');
    if (!menu || !button) return;
    menu.classList.add('hidden');
    button.setAttribute('aria-expanded', 'false');
}

// ===================================
// SIGN-OUT FLOW
// ===================================

// Sign the user out: call POST /auth/logout to revoke the token, then clear
// local state and redirect to the Google sign-in page. On any error the
// local state is still cleared and the redirect still happens, so a failed
// backend request cannot leave the user stranded in an authenticated-looking
// session (design.md Decision 3).
async function signOut() {
    const token = localStorage.getItem('auth_token');
    try {
        if (token) {
            await fetch(`${API_BASE_URL}/auth/logout`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
            });
        }
    } catch (err) {
        console.warn('signOut: server-side logout failed; clearing local state anyway', err);
    } finally {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_email');
        localStorage.removeItem('user_name');
        localStorage.removeItem('user_role');
        currentUser = null;
        window.location.href = '/auth/google/login';
    }
}

// ===================================
// MODAL FOCUS TRAP
// ===================================

// Reusable focus-trap helper for modals. Returns a `release()` function that
// restores focus and unbinds its listeners. Intended to be called when a modal
// opens; the caller MUST invoke the returned release() on close.
//
// Usage:
//   const release = trapFocus(modalEl, () => closeModal());
//   // ... later on close:
//   release();
function trapFocus(modalElement, onClose) {
    if (!modalElement) {
        return () => {};
    }

    const previouslyFocused = document.activeElement;

    const FOCUSABLE_SELECTOR = [
        'a[href]',
        'button:not([disabled])',
        'input:not([disabled]):not([type="hidden"])',
        'select:not([disabled])',
        'textarea:not([disabled])',
        '[tabindex]:not([tabindex="-1"])',
    ].join(',');

    function getFocusable() {
        return Array.from(modalElement.querySelectorAll(FOCUSABLE_SELECTOR))
            .filter(el => el.offsetParent !== null || el === document.activeElement);
    }

    function handleKeydown(e) {
        if (e.key === 'Escape') {
            e.preventDefault();
            if (typeof onClose === 'function') {
                onClose();
            }
            return;
        }
        if (e.key !== 'Tab') return;

        const focusable = getFocusable();
        if (focusable.length === 0) {
            e.preventDefault();
            return;
        }
        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey) {
            if (document.activeElement === first || !modalElement.contains(document.activeElement)) {
                e.preventDefault();
                last.focus();
            }
        } else {
            if (document.activeElement === last) {
                e.preventDefault();
                first.focus();
            }
        }
    }

    modalElement.addEventListener('keydown', handleKeydown);

    // Move initial focus to the first focusable element in the modal.
    const focusable = getFocusable();
    if (focusable.length > 0) {
        focusable[0].focus();
    } else {
        modalElement.setAttribute('tabindex', '-1');
        modalElement.focus();
    }

    return function release() {
        modalElement.removeEventListener('keydown', handleKeydown);
        if (previouslyFocused && typeof previouslyFocused.focus === 'function') {
            previouslyFocused.focus();
        }
    };
}
