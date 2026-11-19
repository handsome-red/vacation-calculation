// Общие элементы интерфейса: уведомления вместо alert, диалог подтверждения вместо
// confirm и защита форм от повторной отправки. Подключается на всех страницах (base.html).
(function () {
    // Уведомление в правом верхнем углу; ошибки висят дольше и читаются скринридером сразу.
    function notify(message, options) {
        const kind = (options && options.kind) || "error";
        let region = document.getElementById("toast-region");
        if (!region) {
            region = document.createElement("div");
            region.id = "toast-region";
            region.className = "toast-region";
            document.body.append(region);
        }

        const toast = document.createElement("div");
        toast.className = `toast toast--${kind}`;
        toast.setAttribute("role", kind === "error" ? "alert" : "status");

        const text = document.createElement("span");
        text.textContent = message;

        const close = document.createElement("button");
        close.type = "button";
        close.className = "toast__close";
        close.setAttribute("aria-label", "Закрыть");
        close.textContent = "✕";
        close.addEventListener("click", () => toast.remove());

        toast.append(text, close);
        region.append(toast);
        setTimeout(() => toast.remove(), kind === "error" ? 10000 : 5000);
    }

    // Диалог подтверждения; возвращает Promise<boolean>.
    function confirmDialog(message, options) {
        const opts = options || {};
        return new Promise((resolve) => {
            const dialog = document.createElement("dialog");
            dialog.className = "confirm-dialog";

            const text = document.createElement("p");
            text.textContent = message;

            const actions = document.createElement("div");
            actions.className = "confirm-dialog__actions";

            const cancel = document.createElement("button");
            cancel.type = "button";
            cancel.className = "btn-light";
            cancel.textContent = opts.cancelLabel || "Отмена";

            const ok = document.createElement("button");
            ok.type = "button";
            ok.className = opts.danger ? "btn btn-danger" : "btn";
            ok.textContent = opts.confirmLabel || "Подтвердить";

            actions.append(cancel, ok);
            dialog.append(text, actions);
            document.body.append(dialog);

            let result = false;
            ok.addEventListener("click", () => { result = true; dialog.close(); });
            cancel.addEventListener("click", () => dialog.close());
            // Esc и закрытие любым способом — отказ.
            dialog.addEventListener("close", () => { dialog.remove(); resolve(result); });

            dialog.showModal();
            (opts.danger ? cancel : ok).focus();
        });
    }

    // <form data-confirm="Текст" data-confirm-label="Удалить" data-danger> — спросить
    // перед отправкой. Подтверждённая форма отправляется тем же кнопкой-отправителем.
    document.addEventListener("submit", async (e) => {
        const form = e.target;
        if (!(form instanceof HTMLFormElement) || !form.dataset.confirm) return;
        if (form.dataset.confirmed === "1") {
            delete form.dataset.confirmed;
            return;
        }
        e.preventDefault();
        e.stopImmediatePropagation();
        const yes = await confirmDialog(form.dataset.confirm, {
            confirmLabel: form.dataset.confirmLabel,
            danger: form.hasAttribute("data-danger"),
        });
        if (!yes) return;
        form.dataset.confirmed = "1";
        form.requestSubmit(e.submitter || undefined);
    }, true);

    // Повторное нажатие «Сохранить», пока идёт запрос, отправило бы форму дважды.
    // Кнопки блокируются после начала отправки: заблокированная до неё не попала бы
    // в данные формы.
    document.addEventListener("submit", (e) => {
        const form = e.target;
        if (!(form instanceof HTMLFormElement) || e.defaultPrevented) return;
        if ((form.method || "get").toLowerCase() !== "post") return;
        setTimeout(() => {
            form.querySelectorAll('button[type="submit"], button:not([type]), input[type="submit"]')
                .forEach((b) => {
                    if (b.disabled) return;
                    b.disabled = true;
                    b.dataset.uiLocked = "1";
                });
        }, 0);
    });

    // Возврат на страницу кнопкой «Назад» восстанавливает её из кэша с заблокированными
    // кнопками: разблокируются только заблокированные здесь.
    window.addEventListener("pageshow", (e) => {
        if (!e.persisted) return;
        document.querySelectorAll("[data-ui-locked]").forEach((b) => {
            b.disabled = false;
            delete b.dataset.uiLocked;
        });
    });

    window.UI = { notify, confirm: confirmDialog };
})();
