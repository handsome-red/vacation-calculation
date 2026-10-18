(function () {
    const el = document.getElementById("calendar");
    if (!el) return;

    const year = parseInt(el.dataset.year, 10) || new Date().getFullYear();

    const MONTHS = ["Январь","Февраль","Март","Апрель","Май","Июнь",
                    "Июль","Август","Сентябрь","Октябрь","Ноябрь","Декабрь"];
    const WEEKDAYS = ["Пн","Вт","Ср","Чт","Пт","Сб","Вс"];

    el.innerHTML = '<p class="muted">Загрузка…</p>';

    fetch(`/api/v1/calendar?year=${year}`, { headers: { Accept: "application/json" } })
        .then((res) => {
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            return res.json();
        })
        .then(render)
        .catch((err) => {
            console.error("calendar load failed:", err);
            el.innerHTML = `<p class="error">Не удалось загрузить календарь: ${err.message}</p>`;
        });

    function render(data) {
        el.innerHTML = "";

        for (const m of data.months) {
            const month = document.createElement("div");
            month.className = "month";

            const h3 = document.createElement("h3");
            h3.textContent = MONTHS[m.month - 1];
            month.appendChild(h3);

            const wdRow = document.createElement("div");
            wdRow.className = "weekdays";
            for (const w of WEEKDAYS) {
                const c = document.createElement("div");
                c.className = "weekday";
                c.textContent = w;
                wdRow.appendChild(c);
            }
            month.appendChild(wdRow);

            const days = document.createElement("div");
            days.className = "days";

            const firstWeekday = new Date(data.year, m.month - 1, 1).getDay();
            const offset = (firstWeekday + 6) % 7;
            for (let k = 0; k < offset; k++) {
                const empty = document.createElement("div");
                empty.className = "day empty";
                days.appendChild(empty);
            }

            for (const d of m.days) {
                const day = document.createElement("div");
                day.className = "day";
                if (d.isWeekend) day.classList.add("weekend");
                if (d.isHoliday) day.classList.add("holiday");
                day.textContent = d.day;
                if (d.holidayName) day.title = d.holidayName;
                days.appendChild(day);
            }

            month.appendChild(days);
            el.appendChild(month);
        }
    }
})();