// API Configuration
const API_BASE_URL = 'http://localhost:8080/api/v1';

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
}

// API Calls
async function apiCall(endpoint, options = {}) {
    try {
        showLoading();
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        });

        if (!response.ok) {
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

// Display Functions
function displaySchedules(schedules) {
    if (!schedules || schedules.length === 0) {
        scheduleList.innerHTML = '<p style="text-align: center; color: #95a5a6; padding: 40px;">No schedules found</p>';
        return;
    }

    scheduleList.innerHTML = schedules.map(schedule => `
        <div class="schedule-card" onclick="loadAndShowDetail('${schedule.id}')">
            <div class="schedule-card-header">
                <div>
                    <div class="schedule-card-title">${escapeHtml(schedule.serviceName)}</div>
                    <div class="schedule-card-id">${schedule.id.substring(0, 8)}</div>
                </div>
                <span class="status-badge status-${schedule.status}">${schedule.status}</span>
            </div>
            <div class="schedule-card-meta">
                <div class="schedule-card-meta-item">
                    <strong>Scheduled</strong>
                    ${formatDateTime(schedule.scheduledAt)}
                </div>
                <div class="schedule-card-meta-item">
                    <strong>Environment</strong>
                    ${schedule.environment}
                </div>
                <div class="schedule-card-meta-item">
                    <strong>Owner</strong>
                    ${escapeHtml(schedule.owner)}
                </div>
            </div>
            ${schedule.description ? `<div style="margin-top: 12px; color: #7f8c8d; font-size: 13px;">${escapeHtml(schedule.description).substring(0, 100)}${schedule.description.length > 100 ? '...' : ''}</div>` : ''}
        </div>
    `).join('');
}

async function loadAndShowDetail(id) {
    const schedule = await getSchedule(id);
    if (schedule) {
        showDetail(schedule);
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
            <div class="detail-label">Environment</div>
            <div class="detail-value">${schedule.environment}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Scheduled Time</div>
            <div class="detail-value">${formatDateTime(schedule.scheduledAt)}</div>
        </div>
        <div class="detail-section">
            <div class="detail-label">Owner</div>
            <div class="detail-value">${escapeHtml(schedule.owner)}</div>
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
    document.getElementById('owner').readOnly = false;
    showView('form');
}

function showEditForm() {
    if (!currentSchedule) return;

    isEditMode = true;
    formTitle.textContent = 'Edit Schedule';

    // Populate form
    document.getElementById('scheduled-at').value = toDateTimeLocalString(currentSchedule.scheduledAt);
    document.getElementById('service').value = currentSchedule.serviceName;
    document.getElementById('environment').value = currentSchedule.environment;
    document.getElementById('owner').value = currentSchedule.owner;
    document.getElementById('owner').readOnly = true; // Owner is immutable
    document.getElementById('description').value = currentSchedule.description || '';
    document.getElementById('rollback-plan').value = currentSchedule.rollbackPlan || '';

    showView('form');
}

async function handleFormSubmit(e) {
    e.preventDefault();

    const scheduledAt = new Date(document.getElementById('scheduled-at').value).toISOString();
    const data = {
        scheduledAt,
        serviceName: document.getElementById('service').value,
        environment: document.getElementById('environment').value,
        owner: document.getElementById('owner').value,
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
