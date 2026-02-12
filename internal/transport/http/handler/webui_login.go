package handler

import (
	"net/http"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// LoginPage serves the login HTML page (GET /web/login).
func (h *Repo) LoginPage(w http.ResponseWriter, r *http.Request) {
	errorParam := r.URL.Query().Get("error")

	errorHTML := ""
	if errorParam == "invalid" {
		errorHTML = `<p style="color: #dc2626; margin-top: 1rem;">Invalid password. Please try again.</p>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Goatway</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .login-container {
            background: white;
            padding: 2.5rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            width: 100%;
            max-width: 400px;
        }
        h1 {
            color: #1f2937;
            font-size: 1.75rem;
            margin-bottom: 0.5rem;
            text-align: center;
        }
        .subtitle {
            color: #6b7280;
            text-align: center;
            margin-bottom: 2rem;
        }
        .form-group {
            margin-bottom: 1.5rem;
        }
        label {
            display: block;
            color: #374151;
            font-weight: 500;
            margin-bottom: 0.5rem;
        }
        input[type="password"] {
            width: 100%;
            padding: 0.75rem 1rem;
            border: 2px solid #e5e7eb;
            border-radius: 8px;
            font-size: 1rem;
            transition: border-color 0.2s;
        }
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 0.875rem;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>🐐 Goatway</h1>
        <p class="subtitle">Admin Dashboard</p>
        <form method="POST" action="/web/login">
            <div class="form-group">
                <label for="password">Admin Password</label>
                <input type="password" id="password" name="password" required
                       placeholder="Enter your admin password" autofocus>
            </div>
            <button type="submit">Sign In</button>
        </form>` + errorHTML + `
    </div>
</body>
</html>`))
}

// Login handles POST /web/login.
func (h *Repo) Login(w http.ResponseWriter, r *http.Request) {
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
func (h *Repo) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("goatway_session")
	if cookie != nil && h.SessionStore != nil {
		h.SessionStore.Delete(cookie.Value)
	}

	auth.ClearSessionCookie(w)

	http.Redirect(w, r, "/web/login", http.StatusFound)
}
