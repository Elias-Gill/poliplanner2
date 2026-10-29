    function scheduleBuilder() {
      return {
        scheduleTitle: "",
        selectedSubjects: [],
        mobileOpen: false,

        init() {
          this.$nextTick(() => {
            if (this.$refs.titleInput) {
              this.$refs.titleInput.focus();
              this.$refs.titleInput.select();
            }
          });
        },

        addSubject(subject) {
          if (!this.selectedSubjects.some((s) => s.id === subject.id)) {
            this.selectedSubjects.push(subject);
          }
        },

        removeSubject(id) {
          this.selectedSubjects = this.selectedSubjects.filter(
            (s) => s.id !== id,
          );
          if (this.selectedSubjects.length === 0) {
            this.mobileOpen = false;
          }
        },

        isSubjectAdded(id) {
          return this.selectedSubjects.some((s) => s.id === id);
        },

        clearAll() {
          this.selectedSubjects = [];
          this.mobileOpen = false;
        },
      };
    }
