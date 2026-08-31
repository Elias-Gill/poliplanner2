document.addEventListener('alpine:init', () => {
    Alpine.data('examScheduleApp', () => ({
        showCalendarModal: false,
        selectedExamId: null,
        calendar: null,
        highlightTimer: null,

        openCalendarModal(examId = null) {
            this.selectedExamId = examId;
            this.showCalendarModal = true;

            this.$nextTick(() => {
                if (!this.calendar) {
                    this.initCalendar();
                } else {
                    this.calendar.updateSize();
                }

                if (examId) {
                    this.applyCalendarEventHighlight(examId);
                    const events = this.calendar.getEvents();
                    const target = events.find(e => e.id === String(examId));
                    if (target && target.start) {
                        this.calendar.gotoDate(target.start);
                    }
                }
            });
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

            let initialDate = undefined;
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

            setTimeout(() => {
                if (this.calendar) {
                    this.calendar.updateSize();
                }
            }, 100);

            // Observador para cambios de modo oscuro/claro
            const observer = new MutationObserver(() => {
                if (this.calendar) {
                    this.calendar.updateSize();
                }
            });
            observer.observe(document.documentElement, {
                attributes: true,
                attributeFilter: ['class']
            });
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
        }
    }));
});
