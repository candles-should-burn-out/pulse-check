<!doctype html>
<html lang="${(locale.currentLanguageTag)!'ru'}">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Восстановление доступа к Pulse Check</title>
    <script src="${url.resourcesPath}/js/theme-mode.js"></script>
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
          <button
            class="theme-toggle"
            type="button"
            aria-label="Переключить тему"
            aria-pressed="false"
            title="Переключить тему"
          >
            <span class="theme-toggle__icon theme-toggle__icon--sun" aria-hidden="true">
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="4" stroke="currentColor" stroke-width="2" />
                <path
                  d="M12 2v2m0 16v2M4 12H2m20 0h-2M5 5l1.4 1.4M17.6 17.6 19 19M19 5l-1.4 1.4M6.4 17.6 5 19"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                />
              </svg>
            </span>
            <span class="theme-toggle__icon theme-toggle__icon--moon" aria-hidden="true">
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none">
                <path
                  d="M20.4 15.2A8.2 8.2 0 0 1 8.8 3.6 8.4 8.4 0 1 0 20.4 15.2Z"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linejoin="round"
                />
              </svg>
            </span>
          </button>
        </div>
      </header>

      <main>
        <section class="shell auth-flow" aria-labelledby="reset-title">
          <div class="login-panel">
            <div class="login-panel__header">
              <div>
                <h1 id="reset-title" class="panel-title">Сброс пароля</h1>
                <p class="panel-subtitle">Укажите логин или email от учетной записи</p>
              </div>
              <span class="panel-badge">OIDC</span>
            </div>

            <#if message?? && message?has_content>
              <div class="alert alert--${message.type}" role="alert">
                ${kcSanitize(message.summary)?no_esc}
              </div>
            </#if>

            <form id="kc-reset-password-form" class="form" action="${url.loginAction}" method="post">
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
                  value="${(auth.attemptedUsername!'')}"
                  autocomplete="username"
                  autofocus
                  aria-invalid="<#if messagesPerField?? && messagesPerField.existsError('username')>true<#else>false</#if>"
                  aria-describedby="<#if messagesPerField?? && messagesPerField.existsError('username')>input-error-username</#if>"
                />
                <#if messagesPerField?? && messagesPerField.existsError('username')>
                  <span id="input-error-username" class="field-error" aria-live="polite">
                    ${kcSanitize(messagesPerField.get('username'))?no_esc}
                  </span>
                </#if>
              </div>

              <div class="form-row form-row--actions">
                <a class="text-link" href="${url.loginUrl}">Вернуться ко входу</a>
                <button class="button" type="submit">Отправить</button>
              </div>
            </form>
          </div>
        </section>
      </main>
    </div>
  </body>
</html>
