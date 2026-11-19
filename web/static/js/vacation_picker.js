(function() {
    const el = document.getElementById("calendar");
    if (!el) return;

    const userId = el.dataset.userId;
    // Администратор может выбрать прошедшие даты (правило обходится с предупреждением).
    const isAdmin = el.dataset.isAdmin === "true";
    const today = isoToday();

    const MONTHS = ["Январь","Февраль","Март","Апрель","Май","Июнь",
                    "Июль","Август","Сентябрь","Октябрь","Ноябрь","Декабрь"];
    const WEEKDAYS = ["Пн","Вт","Ср","Чт","Пт","Сб","Вс"];
    const MONTHS_GENITIVE = ["января","февраля","марта","апреля","мая","июня",
                             "июля","августа","сентября","октября","ноября","декабря"];

    // Сообщение об ошибке на странице (ui.js); без него — обычный alert.
    const notify = (message) => (window.UI ? window.UI.notify(message) : alert(message));

    const state = {
        year: parseInt(el.dataset.year, 10) || new Date().getFullYear(), // показываемый год
        years: new Map(),         // год -> Promise ответа сервера { year, days } (кэш загрузки)
        loaded: new Map(),        // год -> ответ сервера (уже загруженные)
        dayByDate: new Map(),     // "2026-10-01" -> { date, isWeekend, isHoliday, ... } по всем загруженным годам
        busyDates: new Set(),     // "2026-10-01" — дни, занятые другими отпусками
        periods: [],              // [{ from, to, kind, kindLabel }, ...] — могут переходить через год
        pendingStart: null,       // "2026-10-01" — первый клик; второй можно сделать в другом году
        // Сколько дней выбранного вида можно взять с даты start (считает сервер):
        // { start, kind, promise, value: { days, lastDay, unlimited } | null, error }
        limit: null,
    };

    // Переключатель года — над сеткой месяцев.
    const nav = document.createElement("div");
    nav.className = "picker-nav";
    el.before(nav);

    showYear(state.year);

    function loadYear(year) {
        if (!state.years.has(year)) {
            const p = fetch(`/vacations/api/v1/users/${userId}/calendar?year=${year}`, {
                headers: { Accept: "application/json" },
            })
            .then((res) => {
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                return res.json();
            })
            .then((data) => {
                for (const d of data.days) {
                    state.dayByDate.set(d.date, d);
                    if (d.events && d.events.length > 0) {
                        state.busyDates.add(d.date);
                    }
                }
                state.loaded.set(year, data);
                return data;
            })
            .catch((err) => {
                state.years.delete(year); // дать повторить попытку
                throw err;
            });
            state.years.set(year, p);
        }
        return state.years.get(year);
    }

    // Загружает все годы, через которые проходит период [from, to].
    function ensureYears(from, to) {
        const loads = [];
        for (let y = parseInt(from.slice(0, 4), 10); y <= parseInt(to.slice(0, 4), 10); y++) {
            loads.push(loadYear(y));
        }
        return Promise.all(loads);
    }

    async function showYear(year) {
        state.year = year;
        if (!state.loaded.has(year)) {
            renderNav();
            el.innerHTML = '<p class="muted">Загрузка…</p>';
        }
        try {
            await loadYear(year);
        } catch (err) {
            console.error("calendar load failed:", err);
            el.innerHTML = `<p class="error">Не удалось загрузить календарь: ${err.message}</p>`;
            return;
        }
        if (state.year === year) render();
    }

    function render() {
        // Сетка перерисовывается целиком: фокус клавиатуры возвращается на тот же день.
        const focusedDate = el.contains(document.activeElement) ? document.activeElement.dataset.date : null;
        renderNav();
        el.innerHTML = "";
        const data = state.loaded.get(state.year);
        if (!data) return;

        const byMonth = groupByMonth(data.days);

        for (let m = 1; m <= 12; m++) {
            const monthDays = byMonth.get(m) || [];
            if (monthDays.length === 0) continue;

            el.appendChild(renderMonth(data.year, m, monthDays));
        }

        renderPeriods();
        if (focusedDate) {
            const day = el.querySelector(`.day[data-date="${focusedDate}"][tabindex]`);
            if (day) day.focus();
        }
    }

    function renderNav() {
        nav.innerHTML = "";

        const prev = document.createElement("button");
        prev.type = "button";
        prev.textContent = `← ${state.year - 1}`;
        prev.addEventListener("click", () => showYear(state.year - 1));

        const label = document.createElement("strong");
        label.className = "picker-nav__year";
        label.textContent = String(state.year);

        const next = document.createElement("button");
        next.type = "button";
        next.textContent = `${state.year + 1} →`;
        next.addEventListener("click", () => showYear(state.year + 1));

        nav.append(prev, label, next);

        if (state.pendingStart) {
            const hint = document.createElement("span");
            hint.className = "picker-nav__hint";
            hint.textContent = `Начало: ${formatDate(state.pendingStart)} — ${limitHint()}`;
            nav.append(hint);
        }
    }

    function limitHint() {
        const limit = state.limit;
        if (!limit || limit.start !== state.pendingStart) return "выберите дату окончания (можно в другом году)";
        if (limit.error) return `не удалось посчитать доступные дни: ${limit.error}`;
        if (!limit.value) return "считаем доступные дни…";
        if (limit.value.unlimited) return "выберите дату окончания (можно в другом году)";
        if (!limit.value.lastDay) return `с этой даты нет доступных дней вида «${selectedKind().label}» — выберите другой вид или дату`;
        return `вида «${selectedKind().label}» доступно ${limit.value.days} дн., окончание не позже ${formatDate(limit.value.lastDay)}`;
    }

    // Запрашивает у сервера, сколько дней выбранного вида можно взять с даты start
    // вместе с уже выбранными периодами. Ответ кэшируется до смены начала или вида.
    function loadLimit(start) {
        const kind = selectedKind().value;
        if (state.limit && state.limit.start === start && state.limit.kind === kind) {
            return state.limit.promise;
        }
        const limit = { start, kind, value: null, error: null };
        limit.promise = fetch(`/vacations/api/v1/users/${userId}/vacations/limit`, {
            method: "POST",
            headers: { "Content-Type": "application/json", Accept: "application/json" },
            body: JSON.stringify({
                userId,
                kind,
                start,
                periods: state.periods.map((p) => ({ kind: p.kind, startDate: p.from, endDate: p.to })),
            }),
        })
        .then(async (res) => {
            const body = await res.json();
            if (!res.ok) throw new Error(body.error || `HTTP ${res.status}`);
            limit.value = body;
            return body;
        })
        .catch((err) => {
            limit.error = err.message;
            throw err;
        })
        .finally(() => {
            if (state.limit === limit) render();
        });
        state.limit = limit;
        return limit.promise;
    }

    // День за пределами доступных дней вида: окончание на нём не уложится в остаток.
    function isOverLimit(date) {
        const limit = state.limit;
        if (!state.pendingStart || !limit || limit.start !== state.pendingStart || !limit.value) return false;
        if (limit.value.unlimited || date <= state.pendingStart) return false;
        return !limit.value.lastDay || date > limit.value.lastDay;
    }

    function renderMonth(year, monthNumber, days) {
        const month = document.createElement("div");
        month.className = "month";

        const h3 = document.createElement("h3");
        h3.textContent = MONTHS[monthNumber - 1];
        month.appendChild(h3);

        // шапка Пн..Вс
        const wdRow = document.createElement("div");
        wdRow.className = "weekdays";
        for (const w of WEEKDAYS) {
            const c = document.createElement("div");
            c.className = "weekday";
            c.textContent = w;
            wdRow.appendChild(c);
        }
        month.appendChild(wdRow);

        const daysEl = document.createElement("div");
        daysEl.className = "days";

        // offset до 1-го числа (неделя с понедельника)
        const firstWeekday = new Date(year, monthNumber - 1, 1).getDay();
        const offset = (firstWeekday + 6) % 7;
        for (let k = 0; k < offset; k++) {
            const empty = document.createElement("div");
            empty.className = "day empty";
            daysEl.appendChild(empty);
        }

        for (const d of days) {
            daysEl.appendChild(renderDay(d));
        }

        month.appendChild(daysEl);
        return month;
    }

    function renderDay(d) {
        const day = document.createElement("div");
        day.className = "day";
        day.dataset.date = d.date;
        day.textContent = d.day ?? parseInt(d.date.slice(8, 10), 10);

        if (d.isWeekend) day.classList.add("weekend");
        if (d.isHoliday) {
            day.classList.add("holiday");
            if (d.holidayName) day.title = d.holidayName;
        }

        const busy = state.busyDates.has(d.date);
        if (busy) day.classList.add("busy");
        const past = d.date < today;
        if (past) day.classList.add("past");

        if (isInSelected(d.date)) {
            day.classList.add("selected");
        }
        if (state.pendingStart === d.date) day.classList.add("pending");
        const overLimit = isOverLimit(d.date);
        if (overLimit) day.classList.add("over-limit");

        // Отпуск исчисляется в календарных днях: выходные и праздники выбирать можно,
        // нельзя дни, занятые другим отпуском, прошедшие (кроме администратора)
        // и окончание сверх доступных дней выбранного вида.
        const selectable = isSelectable(d.date) && !overLimit;
        day.setAttribute("aria-label", dayLabel(d, { busy, past, overLimit }));
        if (selectable) {
            // Выбор с клавиатуры: Tab до дня, Enter или пробел.
            day.setAttribute("role", "button");
            day.tabIndex = 0;
            day.setAttribute("aria-pressed", String(isInSelected(d.date) || state.pendingStart === d.date));
            day.addEventListener("click", () => onDayClick(d.date));
            day.addEventListener("keydown", (e) => {
                if (e.key !== "Enter" && e.key !== " ") return;
                e.preventDefault();
                onDayClick(d.date);
            });
        } else {
            day.setAttribute("aria-disabled", "true");
        }

        return day;
    }

    // Описание дня для скринридера: «3 июня 2030, суббота, выходной, занят другим отпуском».
    function dayLabel(d, flags) {
        const [y, m, dd] = d.date.split("-").map((s) => parseInt(s, 10));
        const parts = [`${dd} ${MONTHS_GENITIVE[m - 1]} ${y}`];
        if (d.isHoliday) parts.push(d.holidayName ? `праздник: ${d.holidayName}` : "праздник");
        else if (d.isWeekend) parts.push("выходной");
        if (flags.busy) parts.push("занят другим отпуском");
        if (state.pendingStart === d.date) parts.push("выбрано начало");
        else if (isInSelected(d.date)) parts.push("выбран");
        if (flags.past && !isAdmin) parts.push("прошедшая дата");
        if (flags.overLimit) parts.push("сверх доступных дней");
        return parts.join(", ");
    }

    function isInSelected(date) {
        return state.periods.some((p) => date >= p.from && date <= p.to);
    }

    async function onDayClick(date) {
        if (!isSelectable(date)) return;

        if (!state.pendingStart) {
            state.pendingStart = date;
            render();
            loadLimit(date).catch(() => {}); // ошибка показывается в подсказке
            return;
        }

        // Повторный клик по начальной дате отменяет начатый выбор.
        if (date === state.pendingStart) {
            state.pendingStart = null;
            render();
            return;
        }

        const [from, to] = date < state.pendingStart
            ? [date, state.pendingStart]
            : [state.pendingStart, date];

        // Больше доступных дней выбранного вида взять нельзя (при выборе назад
        // начало другое — остаток считается от него).
        const kind = selectedKind();
        let limit;
        try {
            limit = await loadLimit(from);
        } catch (err) {
            notify("Не удалось посчитать доступные дни: " + err.message);
            return;
        }
        if (!limit.unlimited && (!limit.lastDay || to > limit.lastDay)) {
            notify(limit.days > 0
                ? `Вида «${kind.label}» доступно ${limit.days} дн.: с ${formatDate(from)} отпуск можно закончить не позже ${formatDate(limit.lastDay)}`
                : `С ${formatDate(from)} нет доступных дней вида «${kind.label}». Выберите другой вид или другие даты`);
            if (from !== state.pendingStart) loadLimit(state.pendingStart).catch(() => {});
            render();
            return;
        }

        // Период может идти через несколько лет: нужны праздники и занятые дни каждого.
        try {
            await ensureYears(from, to);
        } catch (err) {
            notify("Не удалось загрузить календарь: " + err.message);
            return;
        }

        // блокируем только занятые дни (чужие отпуска).
        const busy = datesBetween(from, to).find((iso) => state.busyDates.has(iso));
        if (busy) {
            notify(`День ${formatDate(busy)} занят другим отпуском`);
            state.pendingStart = null;
            render();
            return;
        }

        // не должно пересекаться с уже выбранными в текущей заявке
        if (state.periods.some((p) => overlaps(p, { from, to }))) {
            notify("Период пересекается с уже выбранным");
            state.pendingStart = null;
            render();
            return;
        }

        state.periods.push({ from, to, kind: kind.value, kindLabel: kind.label });
        state.pendingStart = null;
        state.limit = null; // новые периоды меняют остаток
        render();
        notifyPeriodsChanged();
    }

    function isSelectable(date) {
        return state.dayByDate.has(date) && !state.busyDates.has(date) && (isAdmin || date >= today);
    }

    function isoToday() {
        const d = new Date();
        return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
    }

    function selectedKind() {
        const input = document.querySelector('input[name="vacation-type"]:checked');
        return { value: input?.value || "auto", label: input?.dataset.label || "Автоматически" };
    }

    // Смена вида отпуска применяется к следующим периодам — выбранные не меняются.
    document.querySelectorAll('input[name="vacation-type"]').forEach((input) => {
        input.addEventListener("change", () => {
            document.querySelectorAll(".vacation-type__hint").forEach((hint) => {
                hint.hidden = hint.dataset.for !== selectedKind().value;
            });
            // Начало уже выбрано — остаток пересчитывается для нового вида.
            if (state.pendingStart) {
                loadLimit(state.pendingStart).catch(() => {});
                render();
            }
        });
    });

    function renderPeriods() {
        const box = document.getElementById("selected-periods");
        const btn = document.getElementById("confirm-btn");
        if (!box || !btn) return;

        box.innerHTML = "";

        if (state.periods.length === 0) {
            box.innerHTML = '<p class="muted">Пока ничего не выбрано</p>';
        } else {
            state.periods.forEach((p, i) => {
                const row = document.createElement("div");
                row.className = "period-row";

                const calendarDays = datesBetween(p.from, p.to).length;
                const charged = countChargeableDays(p.from, p.to);
                const note = charged === calendarDays
                    ? `${calendarDays} дн.`
                    : `${calendarDays} дн., из них праздничных ${calendarDays - charged} — не списываются`;
                row.textContent = `${p.kindLabel}: ${formatDate(p.from)} — ${formatDate(p.to)} (${note})`;

                const del = document.createElement("button");
                del.type = "button";
                del.textContent = "✕";
                del.title = "Удалить период";
                del.setAttribute("aria-label", `Удалить период ${formatDate(p.from)} — ${formatDate(p.to)}`);
                del.addEventListener("click", () => {
                    state.periods.splice(i, 1);
                    state.pendingStart = null;
                    state.limit = null;
                    render();
                    notifyPeriodsChanged();
                });
                row.appendChild(del);

                box.appendChild(row);
            });
        }

        // Разблокирует кнопку vacation_form.js — после успешного расчёта распределения.
        if (state.periods.length === 0) btn.disabled = true;
    }

    // Нерабочие праздничные дни в счёт отпуска не входят (ст. 120 ТК РФ).
    function countChargeableDays(from, to) {
        return datesBetween(from, to)
            .filter((iso) => !state.dayByDate.get(iso)?.isHoliday)
            .length;
    }

    function formatDate(iso) {
        const [y, m, d] = iso.split("-");
        return `${d}.${m}.${y}`;
    }

    const btn = document.getElementById("confirm-btn");
    if (btn) {
        btn.addEventListener("click", async () => {
            if (state.periods.length === 0) return;
            btn.disabled = true;

            // Рабочий год (и вид для «Автоматически») определяет сервер при распределении по квотам.
            const payload = {
                userId: userId,
                periods: state.periods.map((p) => ({
                    kind: p.kind,
                    startDate: p.from,
                    endDate: p.to,
                })),
            };

            try {
                const res = await fetch(`/vacations/api/v1/users/${userId}/vacations`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload),
                });

                if (!res.ok) {
                    notify(await res.text());
                    btn.disabled = false;
                    return;
                }

                window.location.href = `/vacations/users/${userId}`;
            } catch (err) {
                notify("Ошибка сети: " + err.message);
                btn.disabled = false;
            }
        });
    }

    function groupByMonth(days) {
        const map = new Map();
        for (const d of days) {
            const m = parseInt(d.date.slice(5, 7), 10);
            if (!map.has(m)) map.set(m, []);
            map.get(m).push(d);
        }
        return map;
    }

    function datesBetween(from, to) {
        const out = [];
        const d = new Date(from + "T00:00:00Z");
        const end = new Date(to + "T00:00:00Z");
        while (d <= end) {
            out.push(d.toISOString().slice(0, 10));
            d.setUTCDate(d.getUTCDate() + 1);
        }
        return out;
    }

    function notifyPeriodsChanged() {
        window.dispatchEvent(new CustomEvent("periods-changed", {
            detail: { userId, periods: state.periods.slice() },
        }));
    }

    function overlaps(a, b) {
        return a.from <= b.to && b.from <= a.to;
    }
}) ();