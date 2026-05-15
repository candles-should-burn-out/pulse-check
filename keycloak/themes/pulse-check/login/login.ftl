<!doctype html>
<html lang="${(locale.currentLanguageTag)!'ru'}">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Вход в Pulse Check</title>
    <link rel="stylesheet" href="${url.resourcesPath}/css/login.css" />
  </head>
  <body>
    <div class="page">
      <header class="topbar">
        <div class="shell topbar__inner">
          <a class="brand" href="${(properties.appUrl)!'http://localhost:3000'}" aria-label="Pulse Check">
            <span class="brand__mark" aria-hidden="true">
              <svg width="19" height="19" viewBox="0 0 24 24" fill="none">
                <path
                  d="M4 12h4l2.2-5 3.6 10 2.2-5h4"
                  stroke="currentColor"
                  stroke-width="2.2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </span>
            <span>Pulse Check</span>
          </a>
        </div>
      </header>

      <main>
        <section class="shell hero" aria-labelledby="login-title">
          <div class="intro">
            <p class="eyebrow">Рабочая область</p>
            <h1 id="login-title">Вход в Pulse Check</h1>
            <p class="lead">
              Авторизуйтесь, чтобы открыть защищенную область и продолжить
              работу с агрегированной статистикой.
            </p>
          </div>

          <div class="login-panel">
            <div class="login-panel__header">
              <div>
                <p class="panel-title">Авторизация</p>
                <p class="panel-subtitle">Введите логин и пароль</p>
              </div>
              <span class="panel-badge">OIDC</span>
            </div>

            <#if message?? && message?has_content>
              <div class="alert alert--${message.type}" role="alert">
                ${kcSanitize(message.summary)?no_esc}
              </div>
            </#if>

            <form id="kc-form-login" class="form" action="${url.loginAction}" method="post">
              <div class="field">
                <label for="username">
                  <#if !realm.loginWithEmailAllowed>
                    Логин
                  <#elseif !realm.registrationEmailAsUsername>
                    Логин или email
                  <#else>
                    Email
                  </#if>
                </label>
                <input
                  id="username"
                  name="username"
                  type="text"
                  value="${login.username!''}"
                  autocomplete="username"
                  autofocus
                  <#if usernameEditDisabled??>disabled</#if>
                />
              </div>

              <div class="field">
                <label for="password">Пароль</label>
                <input
                  id="password"
                  name="password"
                  type="password"
                  autocomplete="current-password"
                />
              </div>

              <#if credentialId??>
                <input type="hidden" name="credentialId" value="${credentialId}" />
              </#if>

              <div class="form-row">
                <#if realm.rememberMe && !(usernameEditDisabled??)>
                  <label class="check">
                    <input
                      id="rememberMe"
                      name="rememberMe"
                      type="checkbox"
                      <#if login.rememberMe??>checked</#if>
                    />
                    <span>Запомнить меня</span>
                  </label>
                </#if>

                <#if realm.resetPasswordAllowed>
                  <a class="text-link" href="${url.loginResetCredentialsUrl}">
                    Забыли пароль?
                  </a>
                </#if>
              </div>

              <button class="button" id="kc-login" name="login" type="submit">
                Войти
              </button>
            </form>
          </div>
        </section>
      </main>
    </div>
  </body>
</html>
