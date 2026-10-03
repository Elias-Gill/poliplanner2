let fullCalendarPromise = null;

// Lazily load the FullCalendar vendor bundle the first time it is needed. It is
// ~280KB, so keeping it out of the initial page load matters on low-end devices.
function loadFullCalendar() {
    if (typeof FullCalendar !== 'undefined') return Promise.resolve();
    if (fullCalendarPromise) return fullCalendarPromise;

    fullCalendarPromise = new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = '/static/vendor/fullcalendar/index.global.min.js';
        script.async = true;
        script.onload = () => resolve();
        script.onerror = () => {
            fullCalendarPromise = null;
            reject(new Error('No se pudo cargar FullCalendar'));
        };
        document.head.appendChild(script);
    });

    return fullCalendarPromise;
}

document.addEventListener('alpine:init', () => {
    Alpine.data('examScheduleApp', () => ({
        showCalendarModal: false,
        selectedExamId: null,
        calendar: null,
        themeObserver: null,
        highlightTimer: null,

        init() {
            // Any close path (buttons, overlay) tears the calendar down so the
            // x-if container can be dropped without leaving stale instances.
            this.$watch('showCalendarModal', (open) => {
                if (!open) this.teardownCalendar();
            });
        },

        async openCalendarModal(examId = null) {
            this.selectedExamId = examId;
            this.showCalendarModal = true;

            await this.$nextTick();

            try {
                await loadFullCalendar();
            } catch (err) {
                console.error(err);
                return;
            }

            this.teardownCalendar();
            this.initCalendar();

            if (examId && this.calendar) {
                this.applyCalendarEventHighlight(examId);
                const events = this.calendar.getEvents();
                const target = events.find((e) => e.id === String(examId));
                if (target && target.start) {
                    this.calendar.gotoDate(target.start);
                }
            }
        },

        initCalendar() {
            const container = this.$refs.calendarContainer;
            if (!container || typeof FullCalendar === 'undefined') return;

            let events = [];
            try {
                const rawData = this.$refs.eventsData.textContent;
                events = JSON.parse(rawData);
            } catch (e) {
                console.error('Error parseando JSON de exámenes:', e);
            }

            let initialDate;
            if (events && events.length > 0) {
                const sorted = [...events].sort((a, b) => new Date(a.start) - new Date(b.start));
                initialDate = sorted[0].start;
            }

            this.calendar = new FullCalendar.Calendar(container, {
                initialView: 'dayGridMonth',
                initialDate: initialDate,
                locale: 'es',
                eventDisplay: 'block',
                displayEventTime: false,
                headerToolbar: {
                    left: 'prev,next',
                    center: 'title',
                    right: 'today'
                },
                height: 'auto',
                events: events,

                eventClassNames: (arg) => {
                    return [`fc-exam-event-${arg.event.id}`];
                },

                eventClick: (info) => {
                    this.selectedExamId = info.event.id;
                    this.showCalendarModal = false;

                    this.applyListHighlight(info.event.id);
                }
            });

            this.calendar.render();

            requestAnimationFrame(() => {
                if (this.calendar) {
                    this.calendar.updateSize();
                }
            });

            this.themeObserver = new MutationObserver(() => {
                if (this.calendar) {
                    this.calendar.updateSize();
                }
            });
            this.themeObserver.observe(document.documentElement, {
                attributes: true,
                attributeFilter: ['class']
            });
        },

        teardownCalendar() {
            if (this.themeObserver) {
                this.themeObserver.disconnect();
                this.themeObserver = null;
            }
            if (this.calendar) {
                this.calendar.destroy();
                this.calendar = null;
            }
        },

        applyCalendarEventHighlight(id) {
            this.$nextTick(() => {
                document.querySelectorAll(".fc-highlighted-event").forEach((el) => {
                    el.classList.remove("fc-highlighted-event");
                });

                const eventEls = document.querySelectorAll(`.fc-exam-event-${id}`);
                eventEls.forEach((el) => {
                    el.classList.add("fc-highlighted-event");
                });
            });
        },

        applyListHighlight(id) {
            this.$nextTick(() => {
                const cardEl = document.getElementById(`exam-card-${id}`);
                if (cardEl) {
                    cardEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }

                if (this.highlightTimer) clearTimeout(this.highlightTimer);
                this.highlightTimer = setTimeout(() => {
                    this.selectedExamId = null;
                }, 2500);
            });
        },

        destroy() {
            if (this.highlightTimer) clearTimeout(this.highlightTimer);
            this.teardownCalendar();
        }
    }));
});
