(function () {
    const widgets = document.querySelectorAll(".calendar-widget");
    if (!widgets.length) return;

    const monthNames = [
        "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
        "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
    ];
    const weekdays = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

    widgets.forEach(initCalendar);

    function initCalendar(el) {
        const userId   = el.dataset.userId;
        const fromName = el.dataset.fromField;
        const toName   = el.dataset.toField;

        if (!userId || !fromName || !toName) {
            console.warn("calendar-widget: missing data attributes", el);
            return;
        }

        let year  = parseInt(el.dataset.year, 10)  || new Date().getFullYear();
        let month = parseInt(el.dataset.month, 10) || (new Date().getMonth() + 1);

        // Скрытые поля для дат — в форме, не внутри виджета.
        const form     = el.closest("form") || document;
        const fromInput = form.querySelector(`input[name="${fromName}"]`);
        const toInput   = form.querySelector(`input[name="${toName}"]`);

        let pickStart = null;   // первая кликнутая дата, ISO-строка
        let data = null;        // ответ сервера на текущий месяц

        el.innerHTML = "";

        const header = document.createElement("div");
        header.className = "cal-header";

        const prev = document.createElement("button");
        prev.type = "button";
        prev.textContent = "←";
        prev.addEventListener("click", () => shiftMonth(-1));

        const label = document.createElement("span");
        label.className = "cal-label";

        const next = document.createElement("button");
        next.type = "button";
        next.textContent = "→";
        next.addEventListener("click", () => shiftMonth(1));

        header.append(prev, label, next);
        el.appendChild(header);

        const grid = document.createElement("div");
        grid.className = "cal-grid";
        el.appendChild(grid);

        // заголовки дней недели
        for (const wd of weekdays) {
            const cell = document.createElement("div");
            cell.className = "cal-weekday";
            cell.textContent = wd;
            grid.appendChild(cell);
        }

        async function load() {
            label.textContent = `${monthNames[month - 1]} ${year}`;

            const url = `/api/v1/users/${userId}/calendar?year=${year}&month=${month}`;
            try {
                const res = await fetch(url);
                if (!res.ok) {
                    throw new Error(`HTTP ${res.status}`);
                }
                data = await res.json();
            } catch (err) {
                console.error("calendar load failed:", err);
                grid.innerHTML = `<p class="error">Не удалось загрузить календарь</p>`;
                return;
            }
            renderGrid();
        }

        function renderGrid() {
            // очищаем всё, кроме заголовков дней недели
            while (grid.children.length > 7) {
                grid.removeChild(grid.lastChild);
            }

            // сдвиг: сколько пустых ячеек до первого дня месяца
            const first = new Date(year, month - 1, 1);
            const offset = (first.getDay() + 6) % 7; // неделя с понедельника
            for (let i = 0; i < offset; i++) {
                const empty = document.createElement("div");
                empty.className = "cal-day cal-empty";
                grid.appendChild(empty);
            }

            // дни месяца
            for (const d of data.days) {
                const cell = document.createElement("button");
                cell.type = "button";
                cell.className = "cal-day";
                cell.dataset.date = d.date;
                cell.textContent = d.day;

                if (d.isWeekend) cell.classList.add("cal-weekend");
                if (d.isHoliday) {
                    cell.classList.add("cal-holiday");
                    if (d.holidayName) cell.title = d.holidayName;
                }
                if (d.events && d.events.length > 0) {
                    cell.classList.add("cal-busy");
                    const kinds = d.events.map(e => e.title).join("; ");
                    cell.title = cell.title ? `${cell.title}; ${kinds}` : kinds;
                    cell.disabled = true; // занятые дни не выбираем
                }

                cell.addEventListener("click", () => onDayClick(d.date));
                grid.appendChild(cell);
            }

            highlightSelection();
        }

        function shiftMonth(delta) {
            month += delta;
            if (month < 1)  { month = 12; year--; }
            if (month > 12) { month = 1;  year++; }
            load();
        }

        function onDayClick(date) {
            if (!fromInput || !toInput) return;

            if (!pickStart) {
                pickStart = date;
                fromInput.value = date;
                toInput.value = date;
                highlightSelection();
                return;
            }

            if (date === pickStart) {
                pickStart = null;
                fromInput.value = "";
                toInput.value = "";
                highlightSelection();
                return;
            }

            const [from, to] = date < pickStart ? [date, pickStart] : [pickStart, date];
            fromInput.value = from;
            toInput.value   = to;
            pickStart = null;
            highlightSelection();
        }

        function highlightSelection() {
            if (!fromInput || !toInput) return;
            const from = fromInput.value;
            const to   = toInput.value;

            grid.querySelectorAll(".cal-day").forEach((cell) => {
                cell.classList.remove("cal-selected");
                const date = cell.dataset.date;
                if (!date) return;

                if (from && to && date >= from && date <= to) {
                    cell.classList.add("cal-selected");
                }
            });
        }

        load();
    }
})();