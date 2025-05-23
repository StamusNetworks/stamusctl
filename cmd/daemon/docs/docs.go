// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/compose/down": {
            "post": {
                "description": "Stops the services defined in the current configuration.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Similar to docker compose down",
                "responses": {
                    "200": {
                        "description": "Down successful",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/compose/init": {
            "post": {
                "description": "Initializes configuration with provided parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Initialize configuration",
                "parameters": [
                    {
                        "description": "Initialization parameters",
                        "name": "arbitraries",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg.InitRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Initialization successful",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/compose/ps": {
            "post": {
                "description": "Will return data about the containers running in the system.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Similar to docker compose ps",
                "responses": {
                    "200": {
                        "description": "List of containers with their status",
                        "schema": {
                            "$ref": "#/definitions/pkg.PsResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/compose/restart/config": {
            "post": {
                "description": "Will restart the containers defined in the current configuration",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Similar to docker restart",
                "responses": {
                    "200": {
                        "description": "Success message",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/compose/restart/containers": {
            "post": {
                "description": "Will restart the containers defined",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Similar to docker restart",
                "parameters": [
                    {
                        "description": "Containers to restart",
                        "name": "arbitraries",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg.Containers"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success message",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/compose/up": {
            "post": {
                "description": "Starts the services defined in the current configuration.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Similar to docker compose up",
                "responses": {
                    "200": {
                        "description": "Up successful",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/compose/update": {
            "post": {
                "description": "Update configuration with provided parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "compose"
                ],
                "summary": "Update configuration",
                "parameters": [
                    {
                        "description": "Update parameters",
                        "name": "arbitraries",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg.UpdateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Initialization successful",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/config": {
            "get": {
                "description": "Retrieves configuration for a given project.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "config"
                ],
                "summary": "Get configuration",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Config name, default is config",
                        "name": "config",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Get content or values, default is false",
                        "name": "content",
                        "in": "query"
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "collectionFormat": "csv",
                        "description": "Values to retrieve, default is all",
                        "name": "values",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Configuration retrieved successfully",
                        "schema": {
                            "$ref": "#/definitions/config.GetResponse"
                        }
                    },
                    "404": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Sets configuration with provided parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "config"
                ],
                "summary": "Set configuration",
                "parameters": [
                    {
                        "description": "Set parameters",
                        "name": "set",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg.SetRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Configuration set successfully",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/config/list": {
            "post": {
                "description": "Sets configuration with provided parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "config"
                ],
                "summary": "Set current configuration",
                "responses": {
                    "200": {
                        "description": "Configuration list",
                        "schema": {
                            "$ref": "#/definitions/pkg.GetListResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "do ping",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "example"
                ],
                "summary": "ping example",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/troubleshoot/containers": {
            "post": {
                "description": "Will return the logs of the containers specified in the request, or all the containers if none are specified.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "logs"
                ],
                "summary": "Logs of the containers",
                "parameters": [
                    {
                        "description": "Logs parameters",
                        "name": "set",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg.LogsRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Containers logs",
                        "schema": {
                            "$ref": "#/definitions/pkg.LogsResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/troubleshoot/kernel": {
            "post": {
                "description": "Will return the logs of the kernel",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "logs"
                ],
                "summary": "Logs of the kernel",
                "responses": {
                    "200": {
                        "description": "Kernel logs",
                        "schema": {
                            "$ref": "#/definitions/pkg.SuccessResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/troubleshoot/reboot": {
            "post": {
                "description": "Will reboot the system",
                "tags": [
                    "reboot"
                ],
                "summary": "Reboots the system",
                "responses": {
                    "500": {
                        "description": "Internal server error with explanation",
                        "schema": {
                            "$ref": "#/definitions/pkg.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/upload": {
            "post": {
                "description": "Handles file uploads",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "upload"
                ],
                "summary": "Upload file example",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Path to save file",
                        "name": "path",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Project name",
                        "name": "project",
                        "in": "query"
                    },
                    {
                        "type": "file",
                        "description": "Upload file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Uploaded file",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error message",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error message",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "config.GetResponse": {
            "type": "object",
            "additionalProperties": true
        },
        "mount.Propagation": {
            "type": "string",
            "enum": [
                "rprivate",
                "private",
                "rshared",
                "shared",
                "rslave",
                "slave"
            ],
            "x-enum-varnames": [
                "PropagationRPrivate",
                "PropagationPrivate",
                "PropagationRShared",
                "PropagationShared",
                "PropagationRSlave",
                "PropagationSlave"
            ]
        },
        "mount.Type": {
            "type": "string",
            "enum": [
                "bind",
                "volume",
                "tmpfs",
                "npipe",
                "cluster"
            ],
            "x-enum-varnames": [
                "TypeBind",
                "TypeVolume",
                "TypeTmpfs",
                "TypeNamedPipe",
                "TypeCluster"
            ]
        },
        "network.EndpointIPAMConfig": {
            "type": "object",
            "properties": {
                "ipv4Address": {
                    "type": "string"
                },
                "ipv6Address": {
                    "type": "string"
                },
                "linkLocalIPs": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "network.EndpointSettings": {
            "type": "object",
            "properties": {
                "aliases": {
                    "description": "Aliases holds the list of extra, user-specified DNS names for this endpoint.",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "dnsnames": {
                    "description": "DNSNames holds all the (non fully qualified) DNS names associated to this endpoint. First entry is used to\ngenerate PTR records.",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "driverOpts": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "endpointID": {
                    "type": "string"
                },
                "gateway": {
                    "type": "string"
                },
                "globalIPv6Address": {
                    "type": "string"
                },
                "globalIPv6PrefixLen": {
                    "type": "integer"
                },
                "ipaddress": {
                    "type": "string"
                },
                "ipamconfig": {
                    "description": "Configurations",
                    "allOf": [
                        {
                            "$ref": "#/definitions/network.EndpointIPAMConfig"
                        }
                    ]
                },
                "ipprefixLen": {
                    "type": "integer"
                },
                "ipv6Gateway": {
                    "type": "string"
                },
                "links": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "macAddress": {
                    "description": "MacAddress may be used to specify a MAC address when the container is created.\nOnce the container is running, it becomes operational data (it may contain a\ngenerated address).",
                    "type": "string"
                },
                "networkID": {
                    "description": "Operational data",
                    "type": "string"
                }
            }
        },
        "pkg.ContainerLogs": {
            "type": "object",
            "properties": {
                "Id": {
                    "type": "string"
                },
                "command": {
                    "type": "string"
                },
                "created": {
                    "type": "integer"
                },
                "hostConfig": {
                    "type": "object",
                    "properties": {
                        "annotations": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        },
                        "networkMode": {
                            "type": "string"
                        }
                    }
                },
                "image": {
                    "type": "string"
                },
                "imageID": {
                    "type": "string"
                },
                "labels": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "logs": {
                    "description": "Logs from the container",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "mounts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/types.MountPoint"
                    }
                },
                "names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "networkSettings": {
                    "$ref": "#/definitions/types.SummaryNetworkSettings"
                },
                "ports": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/types.Port"
                    }
                },
                "sizeRootFs": {
                    "type": "integer"
                },
                "sizeRw": {
                    "type": "integer"
                },
                "state": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "pkg.Containers": {
            "type": "object",
            "properties": {
                "containers": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "pkg.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "pkg.GetListResponse": {
            "type": "object",
            "properties": {
                "configs": {
                    "description": "List of available configurations on the system",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "pkg.InitRequest": {
            "type": "object",
            "properties": {
                "config": {
                    "description": "Config name, default is config",
                    "type": "string"
                },
                "default": {
                    "description": "Use default settings, default is false",
                    "type": "boolean"
                },
                "from_file": {
                    "description": "Values keys and paths to files containing the content used as value",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "project": {
                    "description": "Project name, default is \"clearndr\"",
                    "type": "string"
                },
                "values": {
                    "description": "Values to set, key is the name of the value, value is the value",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "values_path": {
                    "description": "Path to a values.yaml file",
                    "type": "string"
                },
                "version": {
                    "description": "Target version, default is latest",
                    "type": "string"
                }
            }
        },
        "pkg.LogsRequest": {
            "type": "object",
            "properties": {
                "containers": {
                    "description": "Containers ids to show logs from, default is all",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "since": {
                    "description": "Show logs since (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)",
                    "type": "string"
                },
                "tail": {
                    "description": "Number of lines to show from the end, default is all",
                    "type": "string"
                },
                "timestamps": {
                    "description": "Show timestamps, default is false",
                    "type": "boolean"
                },
                "until": {
                    "description": "Show logs until(e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)",
                    "type": "string"
                }
            }
        },
        "pkg.LogsResponse": {
            "type": "object",
            "properties": {
                "containers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/pkg.ContainerLogs"
                    }
                }
            }
        },
        "pkg.PsResponse": {
            "type": "object",
            "properties": {
                "containers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/types.Container"
                    }
                }
            }
        },
        "pkg.SetRequest": {
            "type": "object",
            "properties": {
                "apply": {
                    "description": "Apply the new configuration, relaunch it, default is false",
                    "type": "boolean"
                },
                "config": {
                    "description": "Config name, default is config",
                    "type": "string"
                },
                "from_file": {
                    "description": "Values keys and paths to files containing the content used as value",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "reload": {
                    "description": "Reload the configuration, don't keep arbitrary parameters",
                    "type": "boolean"
                },
                "values": {
                    "description": "Values to set, key is the name of the value, value is the value",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "values_path": {
                    "description": "Path to a values.yaml file",
                    "type": "string"
                }
            }
        },
        "pkg.SuccessResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "pkg.UpdateRequest": {
            "type": "object",
            "properties": {
                "values": {
                    "description": "Values to set, key is the name of the value, value is the value",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "version": {
                    "description": "Version to update to, default is latest",
                    "type": "string"
                }
            }
        },
        "types.Container": {
            "type": "object",
            "properties": {
                "Id": {
                    "type": "string"
                },
                "command": {
                    "type": "string"
                },
                "created": {
                    "type": "integer"
                },
                "hostConfig": {
                    "type": "object",
                    "properties": {
                        "annotations": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        },
                        "networkMode": {
                            "type": "string"
                        }
                    }
                },
                "image": {
                    "type": "string"
                },
                "imageID": {
                    "type": "string"
                },
                "labels": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "mounts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/types.MountPoint"
                    }
                },
                "names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "networkSettings": {
                    "$ref": "#/definitions/types.SummaryNetworkSettings"
                },
                "ports": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/types.Port"
                    }
                },
                "sizeRootFs": {
                    "type": "integer"
                },
                "sizeRw": {
                    "type": "integer"
                },
                "state": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "types.MountPoint": {
            "type": "object",
            "properties": {
                "destination": {
                    "description": "Destination is the path relative to the container root (` + "`" +
	`/` + "`" + `) where the\nSource is mounted inside the container.",
                    "type": "string"
                },
                "driver": {
                    "description": "Driver is the volume driver used to create the volume (if it is a volume).",
                    "type": "string"
                },
                "mode": {
                    "description": "Mode is a comma separated list of options supplied by the user when\ncreating the bind/volume mount.\n\nThe default is platform-specific (` +
	"`" + `\"z\"` + "`" + ` on Linux, empty on Windows).",
                    "type": "string"
                },
                "name": {
                    "description": "Name is the name reference to the underlying data defined by ` +
	"`" + `Source` + "`" + `\ne.g., the volume name.",
                    "type": "string"
                },
                "propagation": {
                    "description": "Propagation describes how mounts are propagated from the host into the\nmount point, and vice-versa. Refer to the Linux kernel documentation\nfor details:\nhttps://www.kernel.org/doc/Documentation/filesystems/sharedsubtree.txt\n\nThis field is not used on Windows.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/mount.Propagation"
                        }
                    ]
                },
                "rw": {
                    "description": "RW indicates whether the mount is mounted writable (read-write).",
                    "type": "boolean"
                },
                "source": {
                    "description": "Source is the source location of the mount.\n\nFor volumes, this contains the storage location of the volume (within\n` +
	"`" + `/var/lib/docker/volumes/` + "`" + `). For bind-mounts, and ` + "`" + `npipe` + "`" +
	`, this contains\nthe source (host) part of the bind-mount. For ` + "`" + `tmpfs` + "`" + ` mount points, this\nfield is empty.",
                    "type": "string"
                },
                "type": {
                    "description": "Type is the type of mount, see ` + "`" + `Type\u003cfoo\u003e` +
	"`" + ` definitions in\ngithub.com/docker/docker/api/types/mount.Type",
                    "allOf": [
                        {
                            "$ref": "#/definitions/mount.Type"
                        }
                    ]
                }
            }
        },
        "types.Port": {
            "type": "object",
            "properties": {
                "IP": {
                    "description": "Host IP address that the container's port is mapped to",
                    "type": "string"
                },
                "PrivatePort": {
                    "description": "Port on the container\nRequired: true",
                    "type": "integer"
                },
                "PublicPort": {
                    "description": "Port exposed on the host",
                    "type": "integer"
                },
                "Type": {
                    "description": "type\nRequired: true",
                    "type": "string"
                }
            }
        },
        "types.SummaryNetworkSettings": {
            "type": "object",
            "properties": {
                "networks": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/network.EndpointSettings"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Swagger Stamusd API",
	Description:      "Stamus daemon server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
