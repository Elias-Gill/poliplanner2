    function excelPanel() {
      return {
        key: "",
        showKey: false,

        types: [
          { value: "schedule", label: "Horario" },
          { value: "lab", label: "Laboratorio" },
        ],
        type: "schedule",
        period: "1",
        date: new Date().toISOString().slice(0, 10),
        url: "",
        name: "",
        fileName: "",
        file: null,

        syncing: false,
        syncMessage: "",
        syncError: false,

        uploading: false,
        uploadMessage: "",
        uploadError: false,

        setFile(event) {
          const selected = event.target.files[0];
          if (!selected) {
            this.clearFile();
            return;
          }
          this.file = selected;
          this.fileName = selected.name;
        },

        clearFile() {
          this.file = null;
          this.fileName = "";
          const input = document.getElementById("fileInput");
          if (input) input.value = "";
        },

        validate() {
          if (!this.key.trim()) {
            this.uploadError = true;
            this.uploadMessage = "Ingresa la clave de autorización.";
            return false;
          }
          if (!this.period) {
            this.uploadError = true;
            this.uploadMessage = "Selecciona un período válido.";
            return false;
          }
          if (!this.date) {
            this.uploadError = true;
            this.uploadMessage = "Selecciona una fecha válida.";
            return false;
          }
          if (!this.url.trim()) {
            this.uploadError = true;
            this.uploadMessage = "El link de descarga es obligatorio.";
            return false;
          }
          return true;
        },

        async runSync() {
          if (!this.key.trim()) {
            this.syncError = true;
            this.syncMessage = "Ingresa la clave de autorización.";
            return;
          }

          this.syncing = true;
          this.syncError = false;
          this.syncMessage = "Procesando…";

          try {
            const res = await fetch("/excel/sync", {
              method: "POST",
              headers: { Authorization: `Bearer ${this.key}` },
            });

            if (res.ok) {
              this.syncMessage = "Sincronización completada con éxito.";
            } else {
              this.syncError = true;
              this.syncMessage = `Error: ${(await res.text()) || "Solicitud fallida"}`;
            }
          } catch (err) {
            this.syncError = true;
            this.syncMessage = `Error de conexión: ${err.message}`;
          } finally {
            this.syncing = false;
          }
        },

        async runUpload() {
          if (!this.validate()) return;

          this.uploading = true;
          this.uploadError = false;
          this.uploadMessage = "Procesando archivo…";

          const formData = new FormData();
          formData.append("type", this.type);
          formData.append("period", this.period);
          formData.append("date", this.date);
          formData.append("downloadUrl", this.url.trim());
          if (this.name.trim()) formData.append("name", this.name.trim());
          if (this.file) formData.append("file", this.file);

          try {
            const res = await fetch("/excel/sync", {
              method: "POST",
              headers: { Authorization: `Bearer ${this.key}` },
              body: formData,
            });

            if (res.ok) {
              this.uploadMessage = "Fuente procesada correctamente.";
              this.clearFile();
              this.name = "";
            } else {
              this.uploadError = true;
              this.uploadMessage = `Error: ${(await res.text()) || "Fallo en la subida"}`;
            }
          } catch (err) {
            this.uploadError = true;
            this.uploadMessage = `Error de conexión: ${err.message}`;
          } finally {
            this.uploading = false;
          }
        },
      };
    }
