package webui

import (
	"net/http"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// LoginPage serves the login HTML page (GET /web/login).
func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	errorParam := r.URL.Query().Get("error")

	errorHTML := ""
	if errorParam == "invalid" {
		errorHTML = `<p>Invalid password. Please try again.</p>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Goatway</title>
</head>
<body>
    <h1>Goatway</h1>
    <p>Admin Dashboard</p>
    <form method="POST" action="/web/login">
        <div>
            <label for="password">Admin Password</label>
            <input type="password" id="password" name="password" required
                   placeholder="Enter your admin password" autofocus>
        </div>
        <button type="submit">Sign In</button>
    </form>` + errorHTML + `
</body>
</html>`))
}

// Login handles POST /web/login.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	hash, err := h.Storage.GetAdminPasswordHash()
	if err != nil || hash == "" {
		http.Error(w, "Server error: admin not configured", http.StatusInternalServerError)
		return
	}

	valid, _ := storage.VerifyPassword(password, hash)
	if !valid {
		http.Redirect(w, r, "/web/login?error=invalid", http.StatusFound)
		return
	}

	if h.SessionStore == nil {
		http.Error(w, "Server error: sessions not configured", http.StatusInternalServerError)
		return
	}

	session := h.SessionStore.Create()
	auth.SetSessionCookie(w, r, session)

	http.Redirect(w, r, "/web", http.StatusFound)
}

// Logout handles POST /web/logout.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("goatway_session")
	if cookie != nil && h.SessionStore != nil {
		h.SessionStore.Delete(cookie.Value)
	}

	auth.ClearSessionCookie(w)

	http.Redirect(w, r, "/web/login", http.StatusFound)
}
