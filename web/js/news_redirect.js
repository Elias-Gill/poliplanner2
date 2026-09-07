// Helpers para cookies
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

async function checkAndRedirectNews() {
  try {
    const res = await fetch('/static/news.json?v=' + Date.now(), {
      cache: 'no-store'
    });
    
    if (!res.ok) return;

    const data = await res.json();
    if (!data || !data.latest || !data.url) return;

    // Validación de fechas
    const now = new Date();
    if (data.valid_from && now < new Date(data.valid_from)) return;
    if (data.valid_until && now > new Date(data.valid_until)) return;

    // Validación de si ya la vio
    const lastSeenVersion = getCookie('poliplanner_seen_news');
    if (!lastSeenVersion || lastSeenVersion !== data.latest) {
      setCookie('poliplanner_seen_news', data.latest, 365);
      window.location.href = data.url;
    }
  } catch (err) {
    console.error('Error verificando news.json:', err);
  }
}

checkAndRedirectNews();
