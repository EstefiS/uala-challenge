# Uala Challenge - Microblogging API

Esta es una API para una plataforma de microblogging simplificada, similar a Twitter. Permite a los usuarios publicar mensajes cortos, seguir a otros usuarios y ver una línea de tiempo (timeline) con los tweets de las personas a las que siguen.

El proyecto está construido utilizando **Go** y sigue los principios de la **Arquitectura Hexagonal (Puertos y Adaptadores)** para mantener una clara separación entre la lógica de negocio y la infraestructura (base de datos, caché, API web).

## ✨ Features (Características)

-   **Publicar Tweets**: Los usuarios pueden publicar mensajes de hasta 280 caracteres.
-   **Seguir Usuarios**: Un usuario puede seguir a otros para ver sus publicaciones.
-   **Timeline Personalizado**: Cada usuario tiene un timeline optimizado para lecturas rápidas que muestra los tweets de los usuarios seguidos.
-   **Caché de Alto Rendimiento**: Usa Redis para cachear las respuestas del timeline, reduciendo la carga en la base de datos.
-   **Documentación Interactiva**: La API está completamente documentada con Swagger/OpenAPI, permitiendo explorar y probar los endpoints fácilmente.
-   **Contenerizado con Docker**: Toda la aplicación y sus dependencias (PostgreSQL, Redis) están contenerizadas para un despliegue y desarrollo consistentes.

## 🛠️ Tech Stack (Tecnologías Utilizadas)

-   **Lenguaje**: Go (v1.24+)
-   **Framework Web**: Gin
-   **Base de Datos**: PostgreSQL
-   **Caché**: Redis
-   **Contenerización**: Docker & Docker Compose
-   **Documentación API**: Swaggo (Swagger/OpenAPI)

## 🚀 Cómo Empezar (Getting Started)

A continuación se detallan los pasos para clonar y ejecutar el proyecto.

### Prerrequisitos

Asegúrate de tener instalado el siguiente software en tu máquina:

-   **Go**: Versión 1.24 o superior.
-   **Docker** y **Docker Compose**.
-   **swag CLI**: La herramienta para generar la documentación de Swagger. Puedes instalarla con:
    ```bash
    go install [github.com/swaggo/swag/cmd/swag@latest](https://github.com/swaggo/swag/cmd/swag@latest)
    ```
    (Asegúrate de que tu `GOPATH/bin` esté en el `PATH` de tu sistema).

### 1. Ejecutar con Docker (Método Recomendado)

Este método levanta la aplicación completa con sus dependencias reales (PostgreSQL y Redis). Es la forma más fiel de simular un entorno de producción.

1.  **Clona el repositorio:**
    ```bash
    git clone [https://github.com/EstefiS/uala-challenge.git](https://github.com/EstefiS/uala-challenge.git)
    cd uala-challenge
    ```

2.  **Crea y levanta los contenedores:**
    Este comando construirá la imagen de la aplicación Go y levantará los tres servicios (`app`, `db`, `cache`). La primera vez, el servicio `db` ejecutará `schema.sql` para crear las tablas.
    ```bash
    docker-compose up --build
    ```

3.  La API estará corriendo en `http://localhost:8080`.

4.  **Para detener la aplicación:**
    Presiona `Control + C` en la terminal o ejecuta `docker-compose down` para detener y eliminar los contenedores.

### 2. Ejecutar Localmente (Modo `dev` con Mocks)

Este método ejecuta la aplicación directamente en tu máquina. No requiere Docker, PostgreSQL ni Redis. Utiliza un **repositorio mock en memoria**, por lo que es extremadamente rápido para desarrollar y probar la lógica de negocio de forma aislada.

1.  Asegúrate de estar en la raíz del proyecto.

2.  **Ejecuta la aplicación:**
    ```bash
    go run ./cmd/server/main.go
    ```

3.  La API estará corriendo en `http://localhost:8080` usando la base de datos simulada.

## 📖 Documentación de la API

Una vez que la aplicación esté corriendo (usando cualquiera de los dos métodos), puedes acceder a la documentación interactiva de la API generada por Swagger.

Abre tu navegador y ve a:
**[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

Desde esta interfaz, podrás ver todos los endpoints, sus parámetros, las respuestas esperadas y probarlos directamente.

> **Nota:** Si realizas cambios en los comentarios de la API (`// @Summary...`), recuerda regenerar la documentación con:
> `swag init -g ./cmd/server/main.go`

## 🔀 Endpoints de la API

Todos los endpoints están bajo el prefijo `/api/v1` y requieren el header `X-User-ID` para identificar al usuario que realiza la petición.

| Método | Ruta                      | Descripción                                                |
| :----- | :------------------------ | :--------------------------------------------------------- |
| `POST` | `/tweets`                 | Publica un nuevo tweet.                                    |
| `POST` | `/users/{id}/follow`      | El usuario actual sigue al usuario con el `{id}` especificado. |
| `GET`  | `/timeline`               | Obtiene el timeline del usuario actual.   