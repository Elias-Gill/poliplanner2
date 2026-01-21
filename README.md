# Poliplanner

Poliplanner es un servicio web para la creación y gestión de horarios académicos.
Está desarrollado íntegramente en **Go**.

## Requisitos

- Go (versión reciente, recomendado Go ≥ 1.21)
- No se requieren bases de datos ni servicios externos para levantar el proyecto (se utiliza
  sqlite3 como BD)

## Configuración

Las variables de entorno están documentadas en el archivo `example.env`.

La **única variable obligatoria** es:

- `UPDATE_KEY`:
  contraseña utilizada para proteger el endpoint de actualización manual del Excel (`/excel`).
  Sin esta variable, el servidor no se levantara y terminara con un mensaje de error.

## Ejecución

Para correr el proyecto en modo desarrollo:

```bash
go run .
```

Para compilar el binario:

```bash
go build
./poliplanner
```

## Detalles técnicos

- El backend está escrito en Go puro.
- Las vistas utilizan `html/template` de la biblioteca estándar.
- El CSS es **semántico**, sin clases:
  los estilos se aplican directamente sobre los tags HTML.
- El JavaScript se escribe directamente dentro de las templates.
- Se utiliza **HTMX** de forma limitada para agregar cierta reactividad sin introducir un
  frontend pesado.

Mas informacion se puede encontrar en la carpeta de `docs`.

### Limitaciones de la plataforma de despliegue

Actualmente el servicio se encuentra desplegado en Fly.io.
Debido al plan gratuito, existen algunas limitaciones operativas que es importante tener en
cuenta.

El servidor entra en auto-suspensión luego de algunos minutos de inactividad.
Esto implica que cualquier dato almacenado únicamente en memoria volátil se pierde cuando la
instancia se apaga.

Además, el despliegue está configurado para consumir la menor cantidad posible de recursos (256
MB de RAM y 1 core compartido), con el objetivo de mantener el servicio funcionando sin costos.
Por este motivo, las decisiones técnicas y futuras funcionalidades deben considerar estas
restricciones.

## Trabajo pendiente / líneas de evolución

Una de las tareas más importantes a corto plazo es modernizar la capa de frontend.
Algunas ideas que están sobre la mesa:

- Migrar de CSS semántico a un framework de estilos (posiblemente **Tailwind CSS v4**, a
  definir).
- Incorporar **Alpine.js** para manejar estados simples en el frontend.
- Reorganizar y simplificar las templates para que trabajen mejor con HTMX y una estructura más
  clara.

> Estas decisiones aún están abiertas a discusión dentro del proyecto.
