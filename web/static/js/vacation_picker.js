(function() {
    const el = document.getElementById("calendar");
    if (!el) return;

    const userId = el.dataset.userId;
    const year = parseInt(el.dataset.year, 10) || new Date().getFullYear();

    const MONTHS = ["Январь","Февраль","Март","Апрель","Май","Июнь",
                    "Июль","Август","Сентябрь","Октябрь","Ноябрь","Декабрь"];
    const WEEKDAYS = ["Пн","Вт","Ср","Чт","Пт","Сб","Вс"];

    const state = {
        data: null,               // ответ сервера { year, days }
        dayByDate: new Map(),     // "2026-10-01" -> { date, isWeekend, isHoliday, ... }
        busyDates: new Set(),     // "2026-10-01" — дни, занятые подтверждёнными отпусками
        periods: [],              // [{ from, to }, ...]
        pendingStart: null,       // "2026-10-01" — первый клик
    };

    el.innerHTML = '<p class="muted">Загрузка…</p>';

    fetch(`/api/v1/users/${userId}/calendar?year=${year}`, {
        headers: { Accept: "application/json" },
    })
    .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
    })
    .then((data) => {
        window.data = data;
        state.data = data;

        for (const d of data.days) {
            state.dayByDate.set(d.date, d);
            if (d.events && d.events.length > 0) {
                state.busyDates.add(d.date);
            }
        }

        render();
    })
    .catch((err) => {
        console.error("calendar load failed:", err);
        el.innerHTML = `<p class="error">Не удалось загрузить календарь: ${err.message}</p>`;
    });

    function render() {
        el.innerHTML = "";
        if (!state.data) return;
    
        const byMonth = groupByMonth(state.data.days);
    
        for (let m = 1; m <= 12; m++) {
            const monthDays = byMonth.get(m) || [];
            if (monthDays.length === 0) continue;
    
            el.appendChild(renderMonth(m, monthDays));
        }
    
        renderPeriods();
    }

    function renderMonth(monthNumber, days) {
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
        const firstWeekday = new Date(state.data.year, monthNumber - 1, 1).getDay();
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

        if (isInSelected(d.date)) {
            day.classList.add("selected");
        }
        if (state.pendingStart === d.date) day.classList.add("pending");

        const selectable = !d.isWeekend && !d.isHoliday && !busy;
        if (selectable) {
            day.addEventListener("click", () => onDayClick(d.date));
        }

        return day;
    }

    function isInSelected(date) {
        return state.periods.some((p) => date >= p.from && date <= p.to);
    }

    function isReserved(date) {
        return state.periods.some((p) => date >= p.from && date <= p.to);
    }

    function onDayClick(date) {
        // нельзя начать с занятого/выходного дня
        if (!isSelectable(date)) return;

        if (!state.pendingStart) {
            state.pendingStart = date;
            render();
            return;
        }

        const [from, to] = date < state.pendingStart
            ? [date, state.pendingStart]
            : [state.pendingStart, date];

        // блокируем только занятые дни (чужие отпуска).
        const busy = datesBetween(from, to).find((iso) => state.busyDates.has(iso));
        if (busy) {
            alert(`День ${busy} занят другим отпуском`);
            state.pendingStart = null;
            render();
            return;
        }

        // не должно пересекаться с уже выбранными в текущей заявке
        if (state.periods.some((p) => overlaps(p, { from, to }))) {
            alert("Период пересекается с уже выбранным");
            state.pendingStart = null;
            render();
            return;
        }

        state.periods.push({ from, to });
        state.pendingStart = null;
        render();
        notifyPeriodsChanged();
    }

    function isSelectable(date) {
        if (state.busyDates.has(date)) return false;
        const info = state.dayByDate.get(date);
        if (!info) return false;
        // выходные/праздники выбирать нельзя как границу, но внутри периода — можно
        return !info.isWeekend && !info.isHoliday;
    }

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

                const workDays = countWorkDays(p.from, p.to);
                row.textContent = `${p.from} — ${p.to} (${workDays} раб. дн.)`;

                const del = document.createElement("button");
                del.type = "button";
                del.textContent = "✕";
                del.title = "Удалить период";
                del.addEventListener("click", () => {
                    state.periods.splice(i, 1);
                    render();
                    notifyPeriodsChanged();
                });
                row.appendChild(del);

                box.appendChild(row);
            });
        }

        btn.disabled = state.periods.length === 0;
    }

    function countWorkDays(from, to) {
        let n = 0;
        for (const iso of datesBetween(from, to)) {
            const info = state.dayByDate.get(iso);
            if (!info) continue;
            if (!info.isWeekend && !info.isHoliday) n++;
        }
        return n;
    }

    const btn = document.getElementById("confirm-btn");
    if (btn) {
        btn.addEventListener("click", async () => {
            if (state.periods.length === 0) return;
            btn.disabled = true;

            const payload = {
                userId: userId,
                periods: state.periods.map((p) => ({
                    kind: "base",
                    startDate: p.from,
                    endDate: p.to,
                })),
            };

            try {
                const res = await fetch(`/api/v1/users/${userId}/vacations`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload),
                });

                if (!res.ok) {
                    alert(await res.text());
                    btn.disabled = false;
                    return;
                }

                window.location.href = `/users/${userId}`;
            } catch (err) {
                alert("Ошибка сети: " + err.message);
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
        const totalWork = state.periods.reduce(
            (sum, p) => sum + countWorkDays(p.from, p.to),
            0
        );
        console.log("periods-changed:", totalWork, state.periods);
        window.dispatchEvent(new CustomEvent("periods-changed", {
            detail: { periods: state.periods.slice(), totalWork },
        }));
    }

    function overlaps(a, b) {
        return a.from <= b.to && b.from <= a.to;
    }
}) ();