import { renderQuotas } from './quota.js';

const dataEl = document.getElementById('app-data');
const APP = JSON.parse(dataEl.textContent);

// #quota-app рендерится только при активном рабочем годе
const root = document.getElementById('quota-app');
const allocationEl = document.getElementById('allocation');
const confirmBtn = document.getElementById('confirm-btn');

if (root) renderQuotas(root, APP);

const KIND_ORDER = ['base', 'irregular', 'seniority'];

// Номер последнего запроса: ответы на устаревшие запросы игнорируются.
let previewSeq = 0;

// Распределение периодов по квотам считает сервер — та же логика, что и при сохранении.
window.addEventListener('periods-changed', async (e) => {
    const { userId, type, periods } = e.detail;
    const seq = ++previewSeq;
    if (confirmBtn) confirmBtn.disabled = true;

    if (periods.length === 0) {
        if (root) renderQuotas(root, APP);
        renderAllocation([]);
        return;
    }

    let body;
    let ok = false;
    try {
        const res = await fetch(`/api/v1/users/${userId}/vacations/preview`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
            body: JSON.stringify({
                userId,
                type,
                periods: periods.map((p) => ({ startDate: p.from, endDate: p.to })),
            }),
        });
        ok = res.ok;
        body = await res.json();
    } catch (err) {
        body = { error: 'Ошибка сети: ' + err.message };
    }
    if (seq !== previewSeq) return;

    if (!ok) {
        if (root) renderQuotas(root, APP);
        renderError(body.error || 'Не удалось рассчитать отпуск');
        if (confirmBtn) confirmBtn.disabled = true;
        return;
    }

    if (root) renderQuotas(root, { ...APP, yearStats: withReserved(APP.yearStats, body.allocations) });
    renderAllocation(body.allocations);
    renderBox('notices', 'Внимание: дни используются позже срока переноса (не позднее 12 месяцев после окончания рабочего года):', body.notices || []);
    renderBox('warnings', 'Правила не соблюдены — отпуск будет оформлен правами администратора:', body.warnings || []);
    if (confirmBtn) confirmBtn.disabled = false;
});

function withReserved(yearStats, allocations) {
    const stats = yearStats.map((y) => {
        const copy = { ...y };
        for (const kind of KIND_ORDER) copy[kind] = { ...y[kind], reserved: 0 };
        return copy;
    });
    for (const a of allocations) {
        if (!KIND_ORDER.includes(a.kind)) continue; // административный отпуск квоту не расходует
        const stat = stats.find((y) => y.start === a.workYearStart);
        if (stat) stat[a.kind].reserved += a.days;
    }
    return stats;
}

function renderAllocation(allocations) {
    if (!allocationEl) return;
    allocationEl.replaceChildren();
    if (allocations.length === 0) return;

    const h = document.createElement('h3');
    h.textContent = 'Списание дней';
    allocationEl.append(h);

    const table = document.createElement('table');
    table.className = 'allocation-table';
    table.innerHTML = '<thead><tr><th>Период</th><th>Вид отпуска</th><th>Рабочий год (с)</th><th>Дней</th></tr></thead>';
    const tbody = document.createElement('tbody');
    for (const a of allocations) {
        const tr = document.createElement('tr');
        for (const text of [
            `${formatDate(a.start)} — ${formatDate(a.end)}`,
            a.kindLabel,
            formatDate(a.workYearStart),
            String(a.days),
        ]) {
            const td = document.createElement('td');
            td.textContent = text;
            tr.append(td);
        }
        tbody.append(tr);
    }
    table.append(tbody);
    allocationEl.append(table);
}

// Блок предупреждений над таблицей списания:
// warnings — правила, которые обходит администратор (сотруднику приходит ошибка);
// notices — не блокируют оформление (срок переноса).
function renderBox(className, titleText, items) {
    if (!allocationEl || items.length === 0) return;
    const box = document.createElement('div');
    box.className = className;
    const title = document.createElement('p');
    title.textContent = titleText;
    const ul = document.createElement('ul');
    for (const w of items) {
        const li = document.createElement('li');
        li.textContent = w;
        ul.append(li);
    }
    box.append(title, ul);
    allocationEl.prepend(box);
}

function renderError(message) {
    if (!allocationEl) return;
    const p = document.createElement('p');
    p.className = 'error';
    p.textContent = message;
    allocationEl.replaceChildren(p);
}

function formatDate(iso) {
    const [y, m, d] = iso.split('-');
    return `${d}.${m}.${y}`;
}
