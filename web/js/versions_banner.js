// Helpers para manejar cookies fácilmente
function setCookie(name, value, days) {
  let expires = "";
  if (days) {
    const date = new Date();
    date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
    expires = "; expires=" + date.toUTCString();
  }
  document.cookie = name + "=" + (value || "") + expires + "; path=/; SameSite=Lax";
}

function getCookie(name) {
  const nameEQ = name + "=";
  const ca = document.cookie.split(';');
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) === ' ') c = c.substring(1, c.length);
    if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
  }
  return null;
}

window.newsApp = function () {
    return {
        showNotice: false,
        latestNews: null,
        allNews: [],
        loading: true,

        async init() {
            try {
                // Se agrega cache-busting con timestamp y cache: 'no-store'
                const res = await fetch('/static/versions.json?v=' + Date.now(), {
                    cache: 'no-store'
                });
                
                if (!res.ok) return;

                const data = await res.json();
                this.allNews = data.entries || [];

                // Buscamos la entrada correspondiente a la última versión
                this.latestNews = this.allNews.find(entry => entry.version === data.latest);

                if (this.latestNews) {
                    const lastSeenVersion = getCookie('poliplanner_seen_version');

                    // Si nunca vio la versión actual o la guardada es diferente a la "latest"
                    if (!lastSeenVersion || lastSeenVersion !== data.latest) {
                        setTimeout(() => {
                            this.showNotice = true;
                        }, 500);
                    }
                }
            } catch (err) {
                console.error('Error cargando versions.json:', err);
            } finally {
                this.loading = false;
            }
        },

        dismissNotice() {
            this.showNotice = false;
            if (this.latestNews) {
                // Guardamos la versión que el usuario acaba de ver
                setCookie('poliplanner_seen_version', this.latestNews.version, 365);
            }
        }
    };
};
