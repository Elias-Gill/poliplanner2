window.examScheduleApp = function () {
    return {
        activeTab: "list",
        selectedExamId: null,
        calendarInstance: null,
        highlightTimer: null,
        events: [],

        init() {
            try {
                if (this.$refs.eventsData) {
                    this.events = JSON.parse(this.$refs.eventsData.textContent);
                }
            } catch (e) {
                console.error("Error al parsear eventos:", e);
                this.events = [];
            }

            this.initCalendar();

            this.$watch("activeTab", (value) => {
                if (value === "calendar" && this.calendarInstance) {
                    this.$nextTick(() => {
                        setTimeout(() => {
                            this.calendarInstance.updateSize();
                            // Resalta evento si venimos de clic en lista
                            if (this.selectedExamId) {
                                this.applyCalendarEventHighlight(this.selectedExamId);
                            }
                        }, 50);
                    });
                }
            });
        },

        initCalendar() {
            const calendarEl = this.$refs.calendarContainer;
            if (!calendarEl || typeof FullCalendar === "undefined") return;

            if (this.calendarInstance) {
                this.calendarInstance.destroy();
                this.calendarInstance = null;
            }

            let initialDate = undefined;
            if (this.events && this.events.length > 0) {
                const sorted = [...this.events].sort(
                    (a, b) => new Date(a.start) - new Date(b.start)
                );
                initialDate = sorted[0].start;
            }

            this.calendarInstance = new FullCalendar.Calendar(calendarEl, {
                initialView: "dayGridMonth",
                initialDate: initialDate,
                locale: "es",
                eventDisplay: "list-item",
                displayEventTime: false,
                headerToolbar: {
                    left: "prev,next",
                    center: "title",
                    right: "today",
                },
                events: this.events,
                height: "auto",

                eventClassNames: (arg) => {
                    return [`fc-exam-event-${arg.event.id}`];
                },

                eventClick: (info) => {
                    this.highlightExam(info.event.id);
                    this.activeTab = "list";

                    this.$nextTick(() => {
                        const cardEl = document.getElementById(
                            `exam-card-${info.event.id}`
                        );
                        if (cardEl) {
                            cardEl.scrollIntoView({
                                behavior: "smooth",
                                block: "center",
                            });
                        }
                    });
                },
            });

            this.calendarInstance.render();

            setTimeout(() => {
                if (this.calendarInstance) {
                    this.calendarInstance.updateSize();
                }
            }, 100);
        },

        highlightExam(id) {
            this.selectedExamId = id;
            this.applyCalendarEventHighlight(id);

            if (this.highlightTimer) clearTimeout(this.highlightTimer);
            this.highlightTimer = setTimeout(() => {
                this.clearHighlights();
            }, 1500);
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

        clearHighlights() {
            this.selectedExamId = null;
            document.querySelectorAll(".fc-highlighted-event").forEach((el) => {
                el.classList.remove("fc-highlighted-event");
            });
        },

        selectExamFromList(id) {
            const ev = this.events.find((e) => e.id === id);
            if (ev && this.calendarInstance) {
                this.calendarInstance.gotoDate(ev.start);
            }
            this.activeTab = "calendar";
            this.highlightExam(id);
        },
    };
};
