import { MouseEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Route, Routes, useLocation, useNavigate } from "react-router-dom";
import AddCircleOutlineIcon from "@mui/icons-material/AddCircleOutline";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import BarChartOutlinedIcon from "@mui/icons-material/BarChartOutlined";
import CheckIcon from "@mui/icons-material/Check";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import DarkModeOutlinedIcon from "@mui/icons-material/DarkModeOutlined";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import FormatListBulletedOutlinedIcon from "@mui/icons-material/FormatListBulletedOutlined";
import LogoutIcon from "@mui/icons-material/Logout";
import PaletteOutlinedIcon from "@mui/icons-material/PaletteOutlined";
import PersonIcon from "@mui/icons-material/Person";
import RefreshIcon from "@mui/icons-material/Refresh";
import SaveOutlinedIcon from "@mui/icons-material/SaveOutlined";
import WbSunnyOutlinedIcon from "@mui/icons-material/WbSunnyOutlined";
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
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";

import { Entity, fetchEntities } from "./api/entities";
import {
  StatusDefinition,
  StatusInput,
  StatusSet,
  createStatus,
  deleteStatus,
  fetchStatusSet,
  updateStatus,
} from "./api/statuses";
import { useAuth } from "./auth/useAuth";
import { useThemeMode } from "./theme-mode-context";

type LoadState = "idle" | "loading" | "success" | "error";

const stateColor: Record<
  string,
  "default" | "primary" | "success" | "warning" | "error"
> = {
  active: "success",
  pending: "warning",
  disabled: "default",
};

const emptyStatusInput: StatusInput = {
  name: "",
  border_color: "#5e81ac",
  background_color: "#eceff4",
  text_color: "#2e3440",
};

const statusPalettes: Array<{
  label: string;
  colors: Pick<StatusInput, "border_color" | "background_color" | "text_color">;
}> = [
  {
    label: "Синий",
    colors: {
      border_color: "#5e81ac",
      background_color: "#e8f0fb",
      text_color: "#24344d",
    },
  },
  {
    label: "Зеленый",
    colors: {
      border_color: "#4f8f5f",
      background_color: "#e7f4ea",
      text_color: "#1f3f29",
    },
  },
  {
    label: "Янтарный",
    colors: {
      border_color: "#c07b24",
      background_color: "#fff2d8",
      text_color: "#4b2d08",
    },
  },
  {
    label: "Красный",
    colors: {
      border_color: "#bf616a",
      background_color: "#fae5e7",
      text_color: "#4d2027",
    },
  },
  {
    label: "Фиолетовый",
    colors: {
      border_color: "#8d6cab",
      background_color: "#f0e8f7",
      text_color: "#392449",
    },
  },
  {
    label: "Серый",
    colors: {
      border_color: "#6b7280",
      background_color: "#f3f4f6",
      text_color: "#1f2937",
    },
  },
];

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

  return (
    normalizedPathname === "/" ||
    normalizedPathname === "/profile" ||
    normalizedPathname === "/statuses" ||
    normalizedPathname === "/statistics/statuses"
  );
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
  const { mode, toggleMode } = useThemeMode();
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
  const isStatusesRoute = location.pathname === "/statuses";
  const isStatusStatisticsRoute = location.pathname === "/statistics/statuses";
  const logoHref = isWorkspaceRoute ? "/" : ".";
  const isDarkTheme = mode === "dark";

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

  const handleNavigate = useCallback(
    (path: string) => {
      navigate(path);
    },
    [navigate]
  );

  const handleToggleTheme = useCallback(() => {
    toggleMode();
    setProfileAnchor(null);
  }, [toggleMode]);

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
          bgcolor: (theme) =>
            theme.palette.mode === "dark"
              ? "rgba(46, 52, 64, 0.88)"
              : "rgba(236, 239, 244, 0.84)",
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
                  borderColor: (theme) =>
                    theme.palette.mode === "dark"
                      ? "rgba(136, 192, 208, 0.32)"
                      : "rgba(94, 129, 172, 0.28)",
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
              <Stack
                direction="row"
                spacing={0.5}
                alignItems="center"
                useFlexGap
                flexWrap="wrap"
              >
                <Button
                  size="small"
                  variant={isWorkspaceRoute ? "contained" : "text"}
                  startIcon={<FormatListBulletedOutlinedIcon />}
                  onClick={() => handleNavigate("/")}
                >
                  Сущности
                </Button>
                <Button
                  size="small"
                  variant={isStatusesRoute ? "contained" : "text"}
                  startIcon={<PaletteOutlinedIcon />}
                  onClick={() => handleNavigate("/statuses")}
                >
                  Статусы
                </Button>
                <Button
                  size="small"
                  variant={isStatusStatisticsRoute ? "contained" : "text"}
                  startIcon={<BarChartOutlinedIcon />}
                  onClick={() => handleNavigate("/statistics/statuses")}
                >
                  Статистика
                </Button>
              </Stack>
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
                <MenuItem onClick={handleToggleTheme}>
                  {isDarkTheme ? (
                    <WbSunnyOutlinedIcon
                      color="action"
                      fontSize="small"
                      sx={{ mr: 1.25 }}
                    />
                  ) : (
                    <DarkModeOutlinedIcon
                      color="action"
                      fontSize="small"
                      sx={{ mr: 1.25 }}
                    />
                  )}
                  {isDarkTheme ? "Светлая тема" : "Темная тема"}
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
          <Route path="statuses" element={<StatusesPage />} />
          <Route path="statistics/statuses" element={<StatusStatisticsPage />} />
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

function StatusesPage() {
  const { getAccessToken } = useAuth();
  const [statusSet, setStatusSet] = useState<StatusSet | null>(null);
  const [loadStatus, setLoadStatus] = useState<LoadState>("idle");
  const [saveStatus, setSaveStatus] = useState<LoadState>("idle");
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<StatusInput>(emptyStatusInput);
  const [editingStatusID, setEditingStatusID] = useState<string | null>(null);

  const isOwner = statusSet?.role === "status_owner";
  const isEditing = editingStatusID !== null;
  const statuses = statusSet?.statuses ?? [];

  const loadStatusSet = useCallback(
    async (signal?: AbortSignal) => {
      setLoadStatus("loading");
      setError(null);

      try {
        const accessToken = await getAccessToken();
        const nextStatusSet = await fetchStatusSet(accessToken, signal);
        setStatusSet(nextStatusSet);
        setLoadStatus("success");
      } catch (loadError) {
        if (signal?.aborted) {
          return;
        }

        setLoadStatus("error");
        setError(
          loadError instanceof Error
            ? loadError.message
            : "Не удалось загрузить статусы"
        );
      }
    },
    [getAccessToken]
  );

  useEffect(() => {
    const controller = new AbortController();
    void Promise.resolve().then(() => loadStatusSet(controller.signal));

    return () => controller.abort();
  }, [loadStatusSet]);

  const resetForm = useCallback(() => {
    setForm(emptyStatusInput);
    setEditingStatusID(null);
  }, []);

  const handleEditStatus = useCallback((status: StatusDefinition) => {
    setForm({
      name: status.name,
      border_color: status.border_color,
      background_color: status.background_color,
      text_color: status.text_color,
    });
    setEditingStatusID(status.id);
  }, []);

  const handleSaveStatus = useCallback(async () => {
    setSaveStatus("loading");
    setError(null);

    try {
      const accessToken = await getAccessToken();
      const savedStatus = editingStatusID
        ? await updateStatus(accessToken, editingStatusID, form)
        : await createStatus(accessToken, form);

      setStatusSet((current) => {
        if (!current) {
          return current;
        }

        const nextStatuses = editingStatusID
          ? current.statuses.map((status) =>
              status.id === savedStatus.id ? savedStatus : status
            )
          : [...current.statuses, savedStatus];

        return { ...current, statuses: nextStatuses };
      });
      resetForm();
      setSaveStatus("success");
    } catch (saveError) {
      setSaveStatus("error");
      setError(
        saveError instanceof Error
          ? saveError.message
          : "Не удалось сохранить статус"
      );
    }
  }, [editingStatusID, form, getAccessToken, resetForm]);

  const handleDeleteStatus = useCallback(
    async (status: StatusDefinition) => {
      if (!window.confirm(`Удалить статус "${status.name}"?`)) {
        return;
      }

      setSaveStatus("loading");
      setError(null);

      try {
        const accessToken = await getAccessToken();
        await deleteStatus(accessToken, status.id);
        setStatusSet((current) =>
          current
            ? {
                ...current,
                statuses: current.statuses.filter((item) => item.id !== status.id),
              }
            : current
        );
        if (editingStatusID === status.id) {
          resetForm();
        }
        setSaveStatus("success");
      } catch (deleteError) {
        setSaveStatus("error");
        setError(
          deleteError instanceof Error
            ? deleteError.message
            : "Не удалось удалить статус"
        );
      }
    },
    [editingStatusID, getAccessToken, resetForm]
  );

  return (
    <Stack spacing={3}>
      {error ? <Alert severity="error">{error}</Alert> : null}
      {statusSet?.role === "assistant" ? (
        <Alert severity="info">Набор статусов управляется владельцем.</Alert>
      ) : null}

      <Paper variant="outlined">
        <Stack spacing={2.5} sx={{ p: 2.5 }}>
          <Stack
            direction={{ xs: "column", md: "row" }}
            spacing={2}
            alignItems={{ xs: "flex-start", md: "center" }}
            justifyContent="space-between"
          >
            <Box>
              <Typography component="h1" variant="h2">
                Статусы
              </Typography>
              <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                {statuses.length > 0
                  ? `В наборе: ${statuses.length}`
                  : "Набор пока пуст"}
              </Typography>
            </Box>
            <Tooltip title="GET /status-set">
              <span>
                <Button
                  variant="outlined"
                  startIcon={
                    loadStatus === "loading" ? (
                      <CircularProgress color="inherit" size={18} />
                    ) : (
                      <RefreshIcon />
                    )
                  }
                  onClick={() => void loadStatusSet()}
                  disabled={loadStatus === "loading"}
                >
                  Обновить
                </Button>
              </span>
            </Tooltip>
          </Stack>

          {isOwner ? (
            <Box
              component="form"
              onSubmit={(event) => {
                event.preventDefault();
                void handleSaveStatus();
              }}
            >
              <Stack spacing={2}>
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  {statusPalettes.map((palette) => (
                    <Tooltip title={palette.label} key={palette.label}>
                      <Button
                        type="button"
                        variant="outlined"
                        aria-label={palette.label}
                        onClick={() =>
                          setForm((current) => ({ ...current, ...palette.colors }))
                        }
                        sx={{
                          minWidth: 44,
                          width: 44,
                          height: 40,
                          p: 0,
                          borderColor: palette.colors.border_color,
                          bgcolor: palette.colors.background_color,
                          color: palette.colors.text_color,
                          "&:hover": {
                            bgcolor: palette.colors.background_color,
                            borderColor: palette.colors.border_color,
                          },
                        }}
                      >
                        <CheckIcon fontSize="small" />
                      </Button>
                    </Tooltip>
                  ))}
                </Stack>

                <Stack
                  direction={{ xs: "column", md: "row" }}
                  spacing={1.5}
                  alignItems={{ xs: "stretch", md: "center" }}
                >
                  <TextField
                    label="Имя"
                    value={form.name}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        name: event.target.value,
                      }))
                    }
                    required
                    fullWidth
                    slotProps={{ htmlInput: { maxLength: 80 } }}
                  />
                  <TextField
                    label="Контур"
                    type="color"
                    value={form.border_color}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        border_color: event.target.value,
                      }))
                    }
                    sx={{ width: { xs: "100%", md: 116 } }}
                  />
                  <TextField
                    label="Фон"
                    type="color"
                    value={form.background_color}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        background_color: event.target.value,
                      }))
                    }
                    sx={{ width: { xs: "100%", md: 116 } }}
                  />
                  <TextField
                    label="Текст"
                    type="color"
                    value={form.text_color}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        text_color: event.target.value,
                      }))
                    }
                    sx={{ width: { xs: "100%", md: 116 } }}
                  />
                  <StatusPreview status={{ ...form, id: "preview" }} />
                </Stack>

                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  <Button
                    type="submit"
                    variant="contained"
                    startIcon={
                      saveStatus === "loading" ? (
                        <CircularProgress color="inherit" size={18} />
                      ) : isEditing ? (
                        <SaveOutlinedIcon />
                      ) : (
                        <AddCircleOutlineIcon />
                      )
                    }
                    disabled={saveStatus === "loading"}
                  >
                    {isEditing ? "Сохранить" : "Создать"}
                  </Button>
                  {isEditing ? (
                    <Button type="button" variant="text" onClick={resetForm}>
                      Отменить
                    </Button>
                  ) : null}
                </Stack>
              </Stack>
            </Box>
          ) : null}
        </Stack>

        <Divider />

        {statuses.length > 0 ? (
          <TableContainer>
            <Table sx={{ minWidth: 700 }} aria-label="Список статусов">
              <TableHead>
                <TableRow>
                  <TableCell>Статус</TableCell>
                  <TableCell>ID</TableCell>
                  {isOwner ? <TableCell width={120}>Действия</TableCell> : null}
                </TableRow>
              </TableHead>
              <TableBody>
                {statuses.map((status) => (
                  <TableRow key={status.id} hover>
                    <TableCell>
                      <StatusPreview status={status} />
                    </TableCell>
                    <TableCell
                      sx={{
                        fontFamily:
                          'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", monospace',
                        wordBreak: "break-word",
                      }}
                    >
                      {status.id}
                    </TableCell>
                    {isOwner ? (
                      <TableCell>
                        <Stack direction="row" spacing={0.5}>
                          <Tooltip title="Редактировать">
                            <IconButton
                              aria-label="Редактировать"
                              onClick={() => handleEditStatus(status)}
                              size="small"
                            >
                              <EditOutlinedIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Удалить">
                            <IconButton
                              aria-label="Удалить"
                              onClick={() => void handleDeleteStatus(status)}
                              size="small"
                            >
                              <DeleteOutlineIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Stack>
                      </TableCell>
                    ) : null}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        ) : (
          <Box sx={{ p: 4, color: "text.secondary" }}>
            <Typography>
              {loadStatus === "loading" ? "Загружаем статусы" : "Статусов пока нет"}
            </Typography>
          </Box>
        )}
      </Paper>
    </Stack>
  );
}

function StatusStatisticsPage() {
  const { getAccessToken } = useAuth();
  const [statusSet, setStatusSet] = useState<StatusSet | null>(null);
  const [loadStatus, setLoadStatus] = useState<LoadState>("idle");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    async function load() {
      setLoadStatus("loading");
      setError(null);

      try {
        const accessToken = await getAccessToken();
        setStatusSet(await fetchStatusSet(accessToken, controller.signal));
        setLoadStatus("success");
      } catch (loadError) {
        if (controller.signal.aborted) {
          return;
        }

        setLoadStatus("error");
        setError(
          loadError instanceof Error
            ? loadError.message
            : "Не удалось загрузить статусы"
        );
      }
    }

    void load();

    return () => controller.abort();
  }, [getAccessToken]);

  const statuses = statusSet?.statuses ?? [];

  return (
    <Stack spacing={3}>
      {error ? <Alert severity="error">{error}</Alert> : null}

      <Paper variant="outlined">
        <Stack spacing={2.5} sx={{ p: 2.5 }}>
          <Box>
            <Typography component="h1" variant="h2">
              Статистика по статусам
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5 }}>
              {statuses.length > 0
                ? `Измерений: ${statuses.length}`
                : "Измерения пока не настроены"}
            </Typography>
          </Box>
        </Stack>
        <Divider />

        {statuses.length > 0 ? (
          <TableContainer>
            <Table sx={{ minWidth: 560 }} aria-label="Статусы для статистики">
              <TableHead>
                <TableRow>
                  <TableCell>Статус</TableCell>
                  <TableCell width={180}>Роль</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {statuses.map((status) => (
                  <TableRow key={status.id} hover>
                    <TableCell>
                      <StatusPreview status={status} />
                    </TableCell>
                    <TableCell>
                      {statusSet?.role === "assistant" ? "Помощник" : "Владелец"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        ) : (
          <Box sx={{ p: 4, color: "text.secondary" }}>
            <Typography>
              {loadStatus === "loading"
                ? "Загружаем статусы"
                : "Агрегаты по статусам пока пусты"}
            </Typography>
          </Box>
        )}
      </Paper>
    </Stack>
  );
}

function StatusPreview({
  status,
}: {
  status:
    | StatusDefinition
    | (StatusInput & {
        id: string;
      });
}) {
  return (
    <Chip
      label={status.name || "Новый статус"}
      variant="outlined"
      sx={{
        maxWidth: "100%",
        bgcolor: status.background_color,
        borderColor: status.border_color,
        color: status.text_color,
        fontWeight: 700,
        "& .MuiChip-label": {
          overflow: "hidden",
          textOverflow: "ellipsis",
        },
      }}
    />
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
              "linear-gradient(135deg, rgba(136, 192, 208, 0.16), rgba(180, 142, 173, 0.14))",
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
