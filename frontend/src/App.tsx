import { useCallback, useMemo, useState } from "react";
import { Route, Routes } from "react-router-dom";
import RefreshIcon from "@mui/icons-material/Refresh";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Container,
  Divider,
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

type LoadState = "idle" | "loading" | "success" | "error";

const stateColor: Record<
  string,
  "default" | "primary" | "success" | "warning" | "error"
> = {
  active: "success",
  pending: "warning",
  disabled: "default",
};

function EntitiesPage() {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [status, setStatus] = useState<LoadState>("idle");
  const [error, setError] = useState<string | null>(null);

  const hasEntities = entities.length > 0;

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
      const nextEntities = await fetchEntities(controller.signal);
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
  }, []);

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <Box
        component="header"
        sx={{
          borderBottom: 1,
          borderColor: "divider",
          bgcolor: "background.paper",
        }}
      >
        <Container maxWidth="lg" sx={{ py: 3 }}>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={2}
            alignItems={{ xs: "flex-start", sm: "center" }}
            justifyContent="space-between"
          >
            <Box>
              <Typography component="h1" variant="h1">
                Pulse Check
              </Typography>
              <Typography color="text.secondary" sx={{ mt: 0.75 }}>
                Проверка списка сущностей из backend API
              </Typography>
            </Box>

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
                  onClick={handleLoadEntities}
                  disabled={status === "loading"}
                >
                  Загрузить сущности
                </Button>
              </span>
            </Tooltip>
          </Stack>
        </Container>
      </Box>

      <Container component="main" maxWidth="lg" sx={{ py: 4 }}>
        <Stack spacing={3}>
          {status === "error" && error ? (
            <Alert severity="error">{error}</Alert>
          ) : null}

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
                  Данные появятся здесь после успешного запроса.
                </Typography>
              </Box>
            )}
          </Paper>
        </Stack>
      </Container>
    </Box>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<EntitiesPage />} />
    </Routes>
  );
}
