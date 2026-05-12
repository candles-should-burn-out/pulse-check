package utils

const OpenAPISchema = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Pulse Check Backend API",
    "version": "0.1.0",
    "description": "Minimal HTTP service stub for entity listing, health checks, and metrics."
  },
  "paths": {
    "/entities": {
      "get": {
        "summary": "List entities",
        "operationId": "listEntities",
        "responses": {
          "200": {
            "description": "Hardcoded entity list.",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Entity"
                  }
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/metrics": {
      "get": {
        "summary": "Prometheus metrics",
        "operationId": "getMetrics",
        "responses": {
          "200": {
            "description": "Prometheus text metrics.",
            "content": {
              "text/plain; version=0.0.4": {
                "schema": {
                  "type": "string",
                  "example": "pulse_check_entity_list_requests_total 1\n"
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/swagger/": {
      "get": {
        "summary": "OpenAPI schema",
        "operationId": "getOpenAPISchema",
        "responses": {
          "200": {
            "description": "OpenAPI schema for this service.",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object"
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/health/live": {
      "get": {
        "summary": "Liveness probe",
        "operationId": "getLiveness",
        "responses": {
          "200": {
            "$ref": "#/components/responses/HealthOK"
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/health/ready": {
      "get": {
        "summary": "Readiness probe",
        "operationId": "getReadiness",
        "responses": {
          "200": {
            "$ref": "#/components/responses/HealthOK"
          },
          "503": {
            "description": "Application is not ready.",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Status"
                },
                "example": {
                  "status": "not_ready"
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/health/startup": {
      "get": {
        "summary": "Startup probe",
        "operationId": "getStartup",
        "responses": {
          "200": {
            "$ref": "#/components/responses/HealthOK"
          },
          "503": {
            "description": "Application is still starting.",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Status"
                },
                "example": {
                  "status": "starting"
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/livez": {
      "get": {
        "summary": "Liveness probe alias",
        "operationId": "getLivenessAlias",
        "responses": {
          "200": {
            "$ref": "#/components/responses/HealthOK"
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/readyz": {
      "get": {
        "summary": "Readiness probe alias",
        "operationId": "getReadinessAlias",
        "responses": {
          "200": {
            "$ref": "#/components/responses/HealthOK"
          },
          "503": {
            "description": "Application is not ready.",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Status"
                },
                "example": {
                  "status": "not_ready"
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    },
    "/startupz": {
      "get": {
        "summary": "Startup probe alias",
        "operationId": "getStartupAlias",
        "responses": {
          "200": {
            "$ref": "#/components/responses/HealthOK"
          },
          "503": {
            "description": "Application is still starting.",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Status"
                },
                "example": {
                  "status": "starting"
                }
              }
            }
          },
          "405": {
            "$ref": "#/components/responses/MethodNotAllowed"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Entity": {
        "type": "object",
        "required": [
          "id",
          "state"
        ],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "example": "2b72d045-d9d7-43ce-9952-59a0f3e35e88"
          },
          "state": {
            "type": "string",
            "example": "active"
          }
        }
      },
      "Status": {
        "type": "object",
        "required": [
          "status"
        ],
        "properties": {
          "status": {
            "type": "string"
          }
        }
      },
      "Error": {
        "type": "object",
        "required": [
          "error"
        ],
        "properties": {
          "error": {
            "type": "string",
            "example": "method_not_allowed"
          }
        }
      }
    },
    "responses": {
      "HealthOK": {
        "description": "Application health check is OK.",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Status"
            },
            "example": {
              "status": "ok"
            }
          }
        }
      },
      "MethodNotAllowed": {
        "description": "HTTP method is not allowed for this endpoint.",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Error"
            }
          }
        }
      }
    }
  }
}
`
