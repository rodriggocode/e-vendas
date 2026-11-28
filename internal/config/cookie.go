package config

import(
    "net/http"
)


func SetCookie(w http.ResponseWriter, token string){
    cookie := http.Cookie{
        Name: "access_token",
        Value: token,
        MaxAge: 3600,
        Path: "/",
        HttpOnly: true,
        Secure: true,
        SameSite: http.SameSiteDefaultMode,
    }

    http.SetCookie(w, &cookie)
}

// ler o cookie
func ReadToken(r *http.Request) (string, error){
    cookie, err := r.Cookie("access_token")
    if err != nil{
        return "", err
    }

    return cookie.Value, nil
}
