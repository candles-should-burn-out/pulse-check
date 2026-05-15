import { MouseEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Route, Routes, useLocation, useNavigate } from "react-router-dom";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import LogoutIcon from "@mui/icons-material/Logout";
import PersonIcon from "@mui/icons-material/Person";
import RefreshIcon from "@mui/icons-material/Refresh";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Chip,
  CircularProgress,
  Container,
  Divider,
  IconButton,
  Link,
  Menu,
  MenuItem,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Typography,
} from "@mui/material";

import { Entity, fetchEntities } from "./api/entities";
import { useAuth } from "./auth/useAuth";

type LoadState = "idle" | "loading" | "success" | "error";

const stateColor: Record<
  string,
  "default" | "primary" | "success" | "warning" | "error"
> = {
  active: "success",
  pending: "warning",
  disabled: "default",
};

function ProtectedRoute() {
  const { status } = useAuth();
  const location = useLocation();
  const isKnownRoute = isKnownAppRoute(location.pathname);

  useEffect(() => {
    if (status === "anonymous" && isKnownRoute) {
      window.location.replace("/");
    }
  }, [isKnownRoute, status]);

  if (!isKnownRoute && status !== "authenticated") {
    return <StandaloneNotFoundPage />;
  }

  if (status === "authenticated") {
    return <AuthenticatedApp />;
  }

  if (status === "error") {
    return (
      <CenteredState
        title="Авторизация недоступна"
        message="Проверьте настройки Keycloak и попробуйте обновить страницу."
      />
    );
  }

  return (
    <CenteredState
      title="Проверяем авторизацию"
      message="Рабочая область доступна только после входа."
      loading
    />
  );
}

function isKnownAppRoute(pathname: string) {
  const normalizedPathname =
    pathname.length > 1 ? pathname.replace(/\/+$/, "") : pathname;

  return normalizedPathname === "/" || normalizedPathname === "/profile";
}

function LoginRedirect() {
  const { status, login } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const handlePageShow = (event: PageTransitionEvent) => {
      if (event.persisted) {
        window.location.replace("/");
      }
    };

    window.addEventListener("pageshow", handlePageShow);

    return () => {
      window.removeEventListener("pageshow", handlePageShow);
    };
  }, []);

  useEffect(() => {
    if (isBackForwardNavigation()) {
      window.location.replace("/");
      return;
    }

    if (status === "authenticated") {
      navigate("/", { replace: true });
      return;
    }

    if (status === "anonymous") {
      window.history.replaceState(window.history.state, "", "/");
      void login().catch(() => {
        window.location.replace("/");
      });
    }
  }, [login, navigate, status]);

  if (status === "error") {
    return (
      <CenteredState
        title="Авторизация недоступна"
        message="Проверьте настройки Keycloak и попробуйте обновить страницу."
      />
    );
  }

  return (
    <CenteredState
      title="Переходим ко входу"
      message="Рабочая область доступна только после авторизации."
      loading
    />
  );
}

function isBackForwardNavigation() {
  const navigationEntry = performance.getEntriesByType("navigation")[0] as
    | PerformanceNavigationTiming
    | undefined;

  return navigationEntry?.type === "back_forward";
}

function AuthenticatedApp() {
  const { getAccessToken, logout, userName } = useAuth();
  const [entities, setEntities] = useState<Entity[]>([]);
  const [status, setStatus] = useState<LoadState>("idle");
  const [error, setError] = useState<string | null>(null);
  const [profileAnchor, setProfileAnchor] = useState<HTMLElement | null>(null);
  const location = useLocation();
  const navigate = useNavigate();

  const hasEntities = entities.length > 0;
  const profileInitial = (userName?.trim().charAt(0) || "U").toUpperCase();
  const displayUserName = userName ?? "user";
  const isProfileMenuOpen = Boolean(profileAnchor);
  const isWorkspaceRoute = location.pathname === "/";
  const logoHref = isWorkspaceRoute ? "/" : ".";

  const stateSummary = useMemo(() => {
    return entities.reduce<Record<string, number>>((acc, entity) => {
      acc[entity.state] = (acc[entity.state] ?? 0) + 1;
      return acc;
    }, {});
  }, [entities]);

  const handleLoadEntities = useCallback(async () => {
    const controller = new AbortController();

    setStatus("loading");
    setError(null);

    try {
      const accessToken = await getAccessToken();
      const nextEntities = await fetchEntities(accessToken, controller.signal);
      setEntities(nextEntities);
      setStatus("success");
    } catch (loadError) {
      setStatus("error");
      setError(
        loadError instanceof Error
          ? loadError.message
          : "Не удалось загрузить сущности"
      );
    }
  }, [getAccessToken]);

  const handleOpenProfileMenu = useCallback(
    (event: MouseEvent<HTMLElement>) => {
      setProfileAnchor(event.currentTarget);
    },
    []
  );

  const handleCloseProfileMenu = useCallback(() => {
    setProfileAnchor(null);
  }, []);

  const handleOpenProfile = useCallback(() => {
    setProfileAnchor(null);
    navigate("/profile");
  }, [navigate]);

  const handleLogoClick = useCallback(
    (event: MouseEvent<HTMLAnchorElement>) => {
      if (isWorkspaceRoute) {
        return;
      }

      event.preventDefault();
      navigate("/");
    },
    [isWorkspaceRoute, navigate]
  );

  const handleLogout = useCallback(() => {
    setProfileAnchor(null);
    void logout();
  }, [logout]);

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <Box
        component="header"
        sx={{
          borderBottom: 1,
          borderColor: "divider",
          bgcolor: "rgba(255, 255, 255, 0.82)",
          backdropFilter: "blur(16px)",
        }}
      >
        <Container maxWidth="lg">
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={2}
            alignItems={{ xs: "stretch", sm: "center" }}
            justifyContent="space-between"
            sx={{ minHeight: 72, py: { xs: 1.5, sm: 0 } }}
          >
            <Box
              component={Link}
              href={logoHref}
              underline="none"
              onClick={handleLogoClick}
              sx={{
                display: "inline-flex",
                alignItems: "center",
                gap: 1.5,
                color: "text.primary",
                width: "fit-content",
                fontSize: "1.12rem",
                fontWeight: 800,
              }}
              aria-label="Pulse Check"
            >
              <Box
                aria-hidden="true"
                component="span"
                sx={{
                  width: 36,
                  height: 36,
                  border: 1,
                  borderColor: "rgba(35, 92, 99, 0.24)",
                  borderRadius: 1,
                  display: "grid",
                  placeItems: "center",
                  bgcolor: "background.paper",
                  color: "primary.main",
                }}
              >
                <Box
                  component="svg"
                  width={19}
                  height={19}
                  viewBox="0 0 24 24"
                  fill="none"
                  sx={{ display: "block" }}
                >
                  <Box
                    component="path"
                    d="M4 12h4l2.2-5 3.6 10 2.2-5h4"
                    stroke="currentColor"
                    strokeWidth={2.2}
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </Box>
              </Box>
              <Typography
                component="span"
                color="inherit"
                sx={{
                  lineHeight: 1,
                }}
              >
                Pulse Check
              </Typography>
            </Box>

            <Stack
              direction="row"
              spacing={1.5}
              alignItems="center"
              justifyContent={{ xs: "space-between", sm: "flex-end" }}
              useFlexGap
              flexWrap="wrap"
            >
              <Tooltip title="Профиль">
                <IconButton
                  aria-label="Профиль"
                  aria-controls={isProfileMenuOpen ? "profile-menu" : undefined}
                  aria-haspopup="menu"
                  aria-expanded={isProfileMenuOpen ? "true" : undefined}
                  onClick={handleOpenProfileMenu}
                  sx={{ p: 0 }}
                >
                  <Avatar
                    sx={{
                      width: 40,
                      height: 40,
                      bgcolor: "primary.main",
                      color: "primary.contrastText",
                      fontSize: "0.95rem",
                      fontWeight: 700,
                    }}
                  >
                    {profileInitial}
                  </Avatar>
                </IconButton>
              </Tooltip>
              <Menu
                id="profile-menu"
                anchorEl={profileAnchor}
                open={isProfileMenuOpen}
                onClose={handleCloseProfileMenu}
                anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
                transformOrigin={{ vertical: "top", horizontal: "right" }}
                sx={{
                  mt: 1,
                }}
              >
                <MenuItem onClick={handleOpenProfile}>
                  <PersonIcon color="action" fontSize="small" sx={{ mr: 1.25 }} />
                  <Typography
                    component="span"
                    fontWeight={700}
                    noWrap
                    title={displayUserName}
                  >
                    {displayUserName}
                  </Typography>
                </MenuItem>
                <Divider />
                <MenuItem onClick={handleLogout}>
                  <LogoutIcon color="action" fontSize="small" sx={{ mr: 1.25 }} />
                  Выйти
                </MenuItem>
              </Menu>
            </Stack>
          </Stack>
        </Container>
      </Box>

      <Container component="main" maxWidth="lg" sx={{ py: 4 }}>
        <Routes>
          <Route
            index
            element={
              <EntitiesPage
                entities={entities}
                error={error}
                hasEntities={hasEntities}
                stateSummary={stateSummary}
                status={status}
                onLoadEntities={handleLoadEntities}
              />
            }
          />
          <Route path="profile" element={<ProfilePage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Container>
    </Box>
  );
}

function EntitiesPage({
  entities,
  error,
  hasEntities,
  stateSummary,
  status,
  onLoadEntities,
}: {
  entities: Entity[];
  error: string | null;
  hasEntities: boolean;
  stateSummary: Record<string, number>;
  status: LoadState;
  onLoadEntities: () => void;
}) {
  return (
    <Stack spacing={3}>
      {status === "error" && error ? <Alert severity="error">{error}</Alert> : null}

      <Paper variant="outlined">
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          alignItems={{ xs: "flex-start", md: "center" }}
          justifyContent="space-between"
          sx={{ p: 2.5 }}
        >
          <Box>
            <Typography component="h2" variant="h2">
              Сущности
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5 }}>
              {hasEntities
                ? `Загружено: ${entities.length}`
                : "Нажмите кнопку, чтобы запросить данные"}
            </Typography>
          </Box>

          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1.5}
            alignItems={{ xs: "stretch", sm: "center" }}
            justifyContent="flex-end"
          >
            {hasEntities ? (
              <Stack
                direction="row"
                spacing={1}
                useFlexGap
                flexWrap="wrap"
                aria-label="Сводка по состояниям"
              >
                {Object.entries(stateSummary).map(([state, count]) => (
                  <Chip
                    key={state}
                    label={`${state}: ${count}`}
                    color={stateColor[state] ?? "primary"}
                    variant="outlined"
                  />
                ))}
              </Stack>
            ) : null}
            <Tooltip title="GET /entities">
              <span>
                <Button
                  variant="contained"
                  startIcon={
                    status === "loading" ? (
                      <CircularProgress color="inherit" size={18} />
                    ) : (
                      <RefreshIcon />
                    )
                  }
                  onClick={onLoadEntities}
                  disabled={status === "loading"}
                >
                  Загрузить сущности
                </Button>
              </span>
            </Tooltip>
          </Stack>
        </Stack>

        <Divider />

        {hasEntities ? (
          <TableContainer>
            <Table sx={{ minWidth: 560 }} aria-label="Список сущностей">
              <TableHead>
                <TableRow>
                  <TableCell>ID</TableCell>
                  <TableCell width={180}>State</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {entities.map((entity) => (
                  <TableRow key={entity.id} hover>
                    <TableCell
                      sx={{
                        fontFamily:
                          'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", monospace',
                        wordBreak: "break-word",
                      }}
                    >
                      {entity.id}
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={entity.state}
                        color={stateColor[entity.state] ?? "primary"}
                        size="small"
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        ) : (
          <Box sx={{ p: 4, color: "text.secondary" }}>
            <Typography>
              У вас пока нет сущностей
            </Typography>
          </Box>
        )}
      </Paper>
    </Stack>
  );
}

function ProfilePage() {
  const navigate = useNavigate();

  const handleBackToApp = useCallback(() => {
    navigate("/");
  }, [navigate]);

  return (
    <Stack spacing={1.5} alignItems="flex-start">
      <Button
        size="small"
        startIcon={<ArrowBackIcon />}
        onClick={handleBackToApp}
        sx={{
          color: "text.secondary",
          fontWeight: 500,
          px: 0.5,
        }}
      >
        К рабочей области
      </Button>

      <Paper variant="outlined" sx={{ width: "100%" }}>
        <Box
          sx={{
            minHeight: 280,
            p: 3,
          }}
        >
          <Stack spacing={3} alignItems="flex-start">
            <Typography component="h1" variant="h2">
              Настройки профиля
            </Typography>
          </Stack>
        </Box>
      </Paper>
    </Stack>
  );
}

function NotFoundPage() {
  const navigate = useNavigate();

  const handleBackToApp = useCallback(() => {
    navigate("/");
  }, [navigate]);

  return (
    <NotFoundPanel onBack={handleBackToApp} actionLabel="К рабочей области" />
  );
}

function StandaloneNotFoundPage() {
  const handleBackToLanding = useCallback(() => {
    window.location.assign("/");
  }, []);

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default", py: 4 }}>
      <Container maxWidth="md">
        <NotFoundPanel onBack={handleBackToLanding} actionLabel="На главную" />
      </Container>
    </Box>
  );
}

function NotFoundPanel({
  actionLabel,
  onBack,
}: {
  actionLabel: string;
  onBack: () => void;
}) {
  return (
    <Paper
      variant="outlined"
      sx={{
        overflow: "hidden",
      }}
    >
      <Box
        sx={{
          minHeight: 360,
          p: { xs: 3, sm: 5 },
          display: "grid",
          placeItems: "center",
          position: "relative",
        }}
      >
        <Box
          aria-hidden="true"
          sx={{
            position: "absolute",
            inset: 0,
            background:
              "linear-gradient(135deg, rgba(35, 92, 99, 0.08), rgba(138, 79, 61, 0.06))",
          }}
        />
        <Stack
          spacing={2.5}
          alignItems="center"
          sx={{
            maxWidth: 520,
            textAlign: "center",
            position: "relative",
          }}
        >
          <Typography
            aria-hidden="true"
            sx={{
              color: "primary.main",
              fontSize: { xs: "4.75rem", sm: "6.5rem" },
              fontWeight: 800,
              lineHeight: 0.9,
            }}
          >
            404
          </Typography>
          <Stack spacing={1}>
            <Typography component="h1" variant="h1">
              Страница не найдена
            </Typography>
            <Typography color="text.secondary">
              Такой страницы нет или она была перемещена.
            </Typography>
          </Stack>
          <Button
            variant="contained"
            startIcon={<ArrowBackIcon />}
            onClick={onBack}
          >
            {actionLabel}
          </Button>
        </Stack>
      </Box>
    </Paper>
  );
}

function CenteredState({
  title,
  message,
  loading = false,
}: {
  title: string;
  message: string;
  loading?: boolean;
}) {
  return (
    <Box
      sx={{
        minHeight: "100vh",
        bgcolor: "background.default",
        display: "grid",
        placeItems: "center",
        px: 2,
      }}
    >
      <Stack spacing={2} alignItems="center" sx={{ textAlign: "center" }}>
        {loading ? <CircularProgress /> : null}
        <Typography component="h1" variant="h1">
          {title}
        </Typography>
        <Typography color="text.secondary">{message}</Typography>
      </Stack>
    </Box>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="login" element={<LoginRedirect />} />
      <Route path="/*" element={<ProtectedRoute />} />
    </Routes>
  );
}
