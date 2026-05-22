# PRD Acotado — AI Product Decomposer CLI

## 1. Nombre provisional

**AI Product Decomposer CLI**

Otros nombres posibles:

- `apd`
- `mapper`
- `specforge`
- `productmap`
- `flowdoc`
- `docsmith`
- `aidoc`
- `specgo`

Para efectos del PRD se usará:

```txt
apd
```

---

## 2. Problema

Cuando una persona quiere iniciar o modificar un proyecto de software, suele enfrentarse a dos extremos:

```txt
Idea vaga → código improvisado
```

O:

```txt
Plantilla pesada de PRD → documentación que no ayuda a construir
```

El problema es que los documentos tradicionales como PRD, BRD o SRS muchas veces están pensados para equipos humanos, reuniones y aprobaciones, no para trabajar con AI generando código.

El usuario necesita una herramienta que lo guíe paso a paso para pensar mejor el producto, descomponerlo y generar archivos útiles para desarrollo asistido por AI.

---

## 3. Objetivo del producto

Crear una CLI en Go que permita generar documentación estructurada, flexible y editable para proyectos de software, cambios funcionales, features, módulos o ideas nuevas.

La herramienta debe guiar al usuario con preguntas claras, ejemplos, ayudas contextuales y secciones opcionales, generando archivos `.md` y opcionalmente `.json` o `.yaml` que puedan servir como contexto para herramientas de AI, agentes de código o desarrollo manual.

---

## 4. Idea central

La herramienta no debe ser solo un generador de PRD.

Debe ser un:

```txt
Gestor CLI de documentación técnica y funcional orientada a AI
```

Una especie de:

```txt
Word + Excel + PowerPoint + Product Discovery + Backlog Builder
```

pero desde terminal, simple, modular y extensible.

---

## 5. Usuarios objetivo

### Usuario principal

Desarrolladores, product builders, analistas funcionales, freelancers o equipos pequeños que necesitan pasar de una idea o solicitud a una especificación clara para desarrollo.

### Casos típicos

```txt
Tengo una idea y quiero convertirla en backlog.
```

```txt
Tengo un sistema existente y necesito documentar un cambio.
```

```txt
Tengo un proceso de negocio y quiero descomponerlo en módulos.
```

```txt
Tengo una reunión con un cliente y quiero ordenar lo que pidió.
```

```txt
Quiero generar contexto limpio para pedirle código a la AI.
```

---

## 6. Principio del producto

La herramienta debe evitar obligar al usuario a llenar documentos enormes.

Debe funcionar bajo esta lógica:

```txt
El usuario elige qué quiere construir.
La herramienta sugiere una ruta.
El usuario puede aceptar, saltar, editar o agregar secciones.
La herramienta genera documentación útil.
```

---

## 7. Tipos de documentos soportados

La herramienta debe permitir crear diferentes tipos de documentación según la necesidad.

### 7.1 Product Decomposition

Para ideas nuevas o productos desde cero.

Incluye:

- problema
- objetivo
- actores
- impacto esperado
- capacidades
- flujos
- entidades
- reglas de negocio
- módulos
- épicas
- historias de usuario
- tareas sugeridas
- prompts para AI

### 7.2 Change Request

Para cambios sobre sistemas existentes.

Incluye:

- sistema afectado
- situación actual
- problema o necesidad
- cambio solicitado
- módulos afectados
- reglas impactadas
- entidades impactadas
- riesgos
- criterios de aceptación
- tareas técnicas sugeridas

### 7.3 Feature Spec

Para una funcionalidad específica.

Incluye:

- descripción de la feature
- usuario beneficiado
- flujo principal
- casos alternos
- reglas
- UI esperada
- API esperada
- datos requeridos
- criterios de aceptación

### 7.4 Bug / Issue Analysis

Para analizar errores o problemas.

Incluye:

- comportamiento esperado
- comportamiento actual
- pasos para reproducir
- posible causa
- impacto
- archivos o módulos relacionados
- solución propuesta
- pruebas requeridas

### 7.5 Technical Task Pack

Para generar tareas técnicas listas para AI.

Incluye:

- contexto
- objetivo técnico
- archivos esperados
- restricciones
- pasos sugeridos
- criterios de terminado
- prompt listo para AI

---

## 8. Flujo general de uso

```bash
apd new
```

La herramienta pregunta:

```txt
¿Qué quieres crear?
1. Producto nuevo
2. Cambio sobre sistema existente
3. Feature específica
4. Bug / Issue
5. Task técnica
6. Documento vacío personalizado
```

Luego:

```txt
Seleccionas una ruta
↓
La CLI te guía con preguntas
↓
Puedes escribir libremente
↓
Puedes pedir ejemplos
↓
Puedes saltar secciones
↓
Puedes agregar secciones propias
↓
Genera .md
↓
Opcionalmente genera .json/.yaml
↓
Opcionalmente genera backlog/tasks/prompts
```

---

## 9. Experiencia esperada

La CLI no debe sentirse como un formulario rígido.

Debe sentirse como un asistente de pensamiento.

Ejemplo:

```txt
Sección: Problema

Ayuda:
Describe qué está pasando actualmente y por qué eso genera fricción.
No hables todavía de la solución. Enfócate en el dolor, la necesidad o el proceso que falla.

Ejemplo:
"Los funcionarios no pueden verificar rápidamente si un documento fue emitido por el sistema, lo que genera llamadas manuales, pérdida de tiempo y riesgo de falsificación."

Escribe tu respuesta:
>
```

El usuario puede escribir:

```txt
/help
/example
/skip
/edit
/back
/add-section
/done
```

---

## 10. Comandos principales

### Crear nuevo documento

```bash
apd new
```

### Crear documento por tipo

```bash
apd new product
apd new change
apd new feature
apd new bug
apd new task
```

### Continuar documento pendiente

```bash
apd resume
```

### Editar documento existente

```bash
apd edit ./docs/product-map.md
```

### Exportar

```bash
apd export --format md
apd export --format json
apd export --format yaml
```

### Generar backlog desde una spec

```bash
apd backlog ./docs/product-map.md
```

### Generar prompts para AI

```bash
apd prompts ./docs/product-map.md
```

### Validar documento

```bash
apd validate ./docs/product-map.md
```

### Listar plantillas

```bash
apd templates
```

### Crear plantilla personalizada

```bash
apd template create
```

---

## 11. Estructura de salida esperada

Por defecto, la herramienta debe generar una carpeta como esta:

```txt
project-docs/
├── 01-product-map.md
├── 02-domain-model.md
├── 03-user-flows.md
├── 04-backlog.md
├── 05-ai-prompts.md
├── spec.json
└── metadata.yaml
```

Para un Change Request:

```txt
change-request-docs/
├── 01-change-request.md
├── 02-impact-analysis.md
├── 03-affected-modules.md
├── 04-acceptance-criteria.md
├── 05-ai-task-pack.md
└── spec.json
```

---

## 12. Documento principal generado

El `.md` debe tener una estructura limpia, no ceremonial.

Ejemplo para producto nuevo:

```md
# Product Decomposition

## 1. Problema

## 2. Objetivo

## 3. Actores

## 4. Impact Mapping

## 5. Capacidades del sistema

## 6. Flujos principales

## 7. Entidades candidatas

## 8. Reglas de negocio

## 9. Módulos sugeridos

## 10. Épicas

## 11. Historias de usuario

## 12. Criterios de aceptación

## 13. Tareas técnicas sugeridas

## 14. Prompts para AI

## 15. Notas abiertas
```

Todas las secciones deben ser opcionales.

---

## 13. Modo guiado flexible

Cada sección debe tener:

```txt
nombre
objetivo
explicación corta
ejemplo
preguntas guía
respuesta del usuario
opción de saltar
opción de editar
opción de agregar notas
```

Ejemplo interno de sección:

```yaml
id: problem
title: Problema
required: true
description: Define qué situación se quiere resolver.
help: No describas todavía la solución. Describe el dolor, fricción o necesidad.
example: Los usuarios no pueden verificar documentos de forma autónoma.
questions:
  - ¿Qué está pasando actualmente?
  - ¿A quién afecta?
  - ¿Qué consecuencia tiene?
output_key: problem_statement
```

---

## 14. Concepto clave: rutas de documentación

La herramienta debe tener rutas predefinidas.

### Ruta: Producto nuevo

```txt
Problema
Objetivo
Actores
Impact Mapping
Capacidades
Flujos
Entidades
Reglas
Módulos
Backlog
Prompts AI
```

### Ruta: Cambio funcional

```txt
Contexto actual
Cambio solicitado
Motivo
Módulos afectados
Reglas afectadas
Flujo actual
Flujo propuesto
Riesgos
Criterios de aceptación
Tasks AI
```

### Ruta: Feature

```txt
Feature
Actor principal
Resultado esperado
Flujo feliz
Flujos alternos
Reglas
UI/API/Datos
Criterios
Tasks
```

### Ruta: Bug

```txt
Problema observado
Resultado esperado
Pasos para reproducir
Impacto
Hipótesis
Módulos afectados
Solución propuesta
Pruebas
```

---

## 15. Funcionalidad de selección de secciones

La herramienta debe permitir seleccionar secciones antes de empezar.

Ejemplo:

```txt
Selecciona las secciones que deseas incluir:

[x] Problema
[x] Objetivo
[x] Actores
[x] Impact Mapping
[ ] BPMN
[x] Capacidades
[x] Reglas de negocio
[x] Entidades
[ ] Arquitectura
[x] Backlog
[x] Prompts para AI
```

Esto permite que el usuario no tenga que llenar cosas que no necesita.

---

## 16. Funcionalidad “modo libre”

Debe existir un modo donde el usuario no siga una plantilla estricta.

```bash
apd new custom
```

La CLI pregunta:

```txt
¿Cómo quieres llamar este documento?
¿Qué secciones quieres agregar?
¿Quieres partir de una plantilla?
```

El usuario puede crear algo como:

```txt
Minuta de reunión
Análisis de proceso
Requerimiento rápido
Prompt pack
Informe técnico
Checklist de despliegue
```

---

## 17. Funcionalidad “AI-ready output”

La herramienta debe generar una sección final optimizada para AI.

Ejemplo:

```md
# AI Context Pack

## Contexto

## Objetivo

## Restricciones

## Entidades

## Reglas de negocio

## Flujos

## Criterios de aceptación

## Tareas solicitadas

## Instrucciones para la AI
No inventes reglas no especificadas.
Pregunta si falta información crítica.
Respeta las entidades definidas.
Genera código modular y testeable.
```

---

## 18. Prompts generados automáticamente

La herramienta puede generar prompts listos para pegar en ChatGPT, Claude, Codex u OpenCode.

Ejemplo:

```md
## Prompt 1 — Crear estructura base

Actúa como desarrollador senior en Go.

Contexto:
...

Objetivo:
...

Tarea:
Crea la estructura inicial del proyecto siguiendo esta arquitectura.

Restricciones:
...

Entrega:
- árbol de carpetas
- archivos base
- explicación breve
```

Otro ejemplo:

```md
## Prompt 2 — Generar entidades

Con base en las entidades candidatas y reglas de negocio descritas, genera las estructuras en Go, validaciones y tests unitarios.
```

---

## 19. Backlog generado

La herramienta debe poder generar:

```txt
Épicas
Features
Historias de usuario
Criterios de aceptación
Tasks técnicas
```

Ejemplo:

```md
# Backlog

## Épica: Gestión de documentos verificables

### Historia: Verificar documento por QR

Como ciudadano,
quiero escanear un QR,
para confirmar si un documento es válido.

#### Criterios de aceptación

- Dado un QR válido, cuando el usuario lo escanea, entonces el sistema muestra los datos públicos del documento.
- Dado un QR expirado, cuando el usuario lo escanea, entonces el sistema muestra un mensaje de vigencia expirada.
- Dado un documento anulado, cuando el usuario lo escanea, entonces el sistema indica que no es válido.

#### Tasks técnicas

- Crear endpoint público de verificación.
- Crear modelo VerificationToken.
- Crear servicio de validación.
- Crear pantalla pública de resultado.
- Agregar pruebas unitarias.
```

---

## 20. Reglas funcionales

### RF-01 — Crear documento guiado

La CLI debe permitir iniciar un nuevo documento usando una ruta predefinida.

### RF-02 — Seleccionar tipo de documento

La CLI debe permitir seleccionar entre producto nuevo, cambio, feature, bug, task o custom.

### RF-03 — Mostrar ayuda contextual

Cada sección debe mostrar una explicación breve, preguntas guía y ejemplo.

### RF-04 — Permitir saltar secciones

El usuario debe poder omitir secciones sin romper el documento.

### RF-05 — Permitir editar respuestas

El usuario debe poder volver a una sección anterior y modificarla.

### RF-06 — Guardado incremental

La herramienta debe guardar el progreso automáticamente.

### RF-07 — Exportar Markdown

La herramienta debe generar archivos `.md` limpios y editables.

### RF-08 — Exportar JSON/YAML

La herramienta debe generar una representación estructurada para uso posterior.

### RF-09 — Generar backlog

La herramienta debe convertir respuestas en épicas, historias y tareas sugeridas.

### RF-10 — Generar prompts AI

La herramienta debe generar prompts basados en el documento.

### RF-11 — Plantillas configurables

El usuario debe poder crear y reutilizar plantillas propias.

### RF-12 — Secciones personalizadas

El usuario debe poder agregar nuevas secciones no contempladas por defecto.

---

## 21. Reglas no funcionales

### RNF-01 — Simplicidad

La CLI debe ser rápida y fácil de usar desde terminal.

### RNF-02 — Archivos legibles

Todo archivo generado debe poder editarse manualmente.

### RNF-03 — Sin dependencia obligatoria de AI

La herramienta debe funcionar sin conexión a modelos AI.

### RNF-04 — Extensible

Las plantillas deben poder crecer sin recompilar la aplicación.

### RNF-05 — Multiplataforma

Debe funcionar en macOS, Linux y Windows.

### RNF-06 — Bajo acoplamiento

El motor de preguntas, plantillas, exportadores y generadores deben estar separados.

---

## 22. Arquitectura inicial sugerida en Go

```txt
apd/
├── cmd/
│   ├── root.go
│   ├── new.go
│   ├── resume.go
│   ├── edit.go
│   ├── export.go
│   ├── backlog.go
│   └── prompts.go
│
├── internal/
│   ├── app/
│   │   └── application.go
│   │
│   ├── cli/
│   │   ├── prompt.go
│   │   ├── menu.go
│   │   └── renderer.go
│   │
│   ├── templates/
│   │   ├── loader.go
│   │   ├── registry.go
│   │   └── schema.go
│   │
│   ├── document/
│   │   ├── document.go
│   │   ├── section.go
│   │   ├── answer.go
│   │   └── metadata.go
│   │
│   ├── generator/
│   │   ├── markdown.go
│   │   ├── json.go
│   │   ├── yaml.go
│   │   ├── backlog.go
│   │   └── prompts.go
│   │
│   ├── storage/
│   │   ├── filesystem.go
│   │   └── session.go
│   │
│   └── validator/
│       └── validator.go
│
├── templates/
│   ├── product.yaml
│   ├── change-request.yaml
│   ├── feature.yaml
│   ├── bug.yaml
│   └── task.yaml
│
├── examples/
│   └── verification-system/
│
├── go.mod
└── main.go
```

---

## 23. Librerías sugeridas en Go

Para CLI:

```txt
cobra
```

Para prompts interactivos:

```txt
huh
```

O:

```txt
survey
```

Para estilos en terminal:

```txt
lipgloss
```

Para TUI más avanzada:

```txt
bubbletea
```

Para YAML:

```txt
gopkg.in/yaml.v3
```

Para Markdown:

```txt
strings.Builder
```

Inicialmente no necesitas nada complejo.

---

## 24. Modelo de datos inicial

```go
type Document struct {
    ID          string
    Title       string
    Type        string
    Description string
    Sections    []Section
    Metadata    Metadata
}

type Section struct {
    ID          string
    Title       string
    Description string
    Help        string
    Example     string
    Questions   []string
    Required    bool
    Answer      string
    Skipped     bool
}

type Template struct {
    ID          string
    Name        string
    Description string
    Sections    []Section
}

type Metadata struct {
    CreatedAt string
    UpdatedAt string
    Version   string
}
```

---

## 25. MVP recomendado

Para no complicar el desarrollo, el MVP debe incluir solo esto:

```txt
1. Comando apd new
2. Selección de tipo de documento
3. Carga de plantilla YAML
4. Preguntas guiadas
5. Comandos /help, /skip, /back, /done
6. Guardado automático de sesión
7. Exportación a Markdown
8. Generación básica de AI Context Pack
```

No incluir todavía:

```txt
- base de datos
- servidor web
- sincronización cloud
- editor visual
- AI integrada
- colaboración
```

Primero debe generar buenos `.md`.

---

## 26. Roadmap por fases

### Fase 1 — CLI básica

Objetivo:

```txt
Crear documentos guiados desde terminal.
```

Incluye:

- `apd new`
- plantillas YAML
- preguntas interactivas
- exportación Markdown
- guardado local

### Fase 2 — Backlog Generator

Objetivo:

```txt
Transformar la documentación en backlog accionable.
```

Incluye:

- épicas
- historias de usuario
- criterios de aceptación
- tasks técnicas

### Fase 3 — AI Prompt Pack

Objetivo:

```txt
Generar prompts listos para usar con AI.
```

Incluye:

- prompt para arquitectura
- prompt para modelos
- prompt para API
- prompt para UI
- prompt para tests
- prompt para refactor

### Fase 4 — Templates personalizados

Objetivo:

```txt
Permitir que el usuario cree sus propias rutas documentales.
```

Incluye:

- crear plantilla
- editar plantilla
- importar/exportar plantilla
- reutilizar plantilla

### Fase 5 — Modo TUI

Objetivo:

```txt
Convertir la CLI en una experiencia más parecida a una app en terminal.
```

Incluye:

- navegación con teclado
- panel de secciones
- vista previa Markdown
- progreso visual

---

## 27. Criterios de aceptación del MVP

El MVP estará listo cuando:

```txt
- El usuario pueda ejecutar apd new.
- El usuario pueda elegir tipo de documento.
- La CLI cargue una plantilla YAML.
- La CLI muestre ayuda por sección.
- El usuario pueda escribir respuestas largas.
- El usuario pueda saltar secciones.
- El usuario pueda regresar a una sección anterior.
- La CLI guarde progreso automáticamente.
- La CLI genere un archivo Markdown limpio.
- La CLI genere una sección AI Context Pack.
```

---

## 28. Primer ejemplo de plantilla YAML

```yaml
id: product
name: Product Decomposition
description: Ruta para descomponer una idea de producto en contexto útil para desarrollo con AI.

sections:
  - id: problem
    title: Problema
    required: true
    description: Describe el problema principal que se quiere resolver.
    help: No hables todavía de la solución. Explica qué situación genera dolor, fricción o pérdida.
    example: Los usuarios no pueden verificar si un documento fue emitido oficialmente por el sistema.
    questions:
      - ¿Qué está pasando actualmente?
      - ¿A quién afecta?
      - ¿Qué consecuencia tiene?

  - id: goal
    title: Objetivo
    required: true
    description: Define qué se quiere lograr.
    help: Describe el resultado esperado, no una lista de funcionalidades.
    example: Permitir que cualquier ciudadano pueda verificar la validez de un documento mediante QR.
    questions:
      - ¿Qué resultado final quieres conseguir?
      - ¿Cómo sabrás que se resolvió el problema?

  - id: actors
    title: Actores
    required: false
    description: Identifica personas, sistemas o roles involucrados.
    help: Incluye usuarios humanos, sistemas externos, administradores o APIs.
    example: Ciudadano, funcionario, administrador, sistema de documentos.
    questions:
      - ¿Quién usa el sistema?
      - ¿Quién administra?
      - ¿Qué sistemas externos participan?

  - id: capabilities
    title: Capacidades
    required: false
    description: Lista lo que el sistema debe ser capaz de hacer.
    help: Escríbelo como capacidades, no como tareas técnicas.
    example: Generar QR único, validar documento, mostrar vigencia, descargar PDF.
    questions:
      - ¿Qué debe permitir hacer el sistema?
      - ¿Qué capacidades son críticas para el MVP?

  - id: rules
    title: Reglas de negocio
    required: false
    description: Define condiciones, límites y comportamientos obligatorios.
    help: Una regla debe poder evaluarse como verdadera o falsa.
    example: Un documento anulado no debe mostrarse como válido.
    questions:
      - ¿Qué condiciones deben cumplirse?
      - ¿Qué casos deben bloquearse?
      - ¿Qué validaciones son obligatorias?

  - id: ai_context
    title: Contexto para AI
    required: false
    description: Resume restricciones importantes para generar código.
    help: Aquí debes indicar stack, arquitectura, estilo de código o reglas que la AI debe respetar.
    example: Usar Go, arquitectura limpia, archivos pequeños, tests unitarios.
    questions:
      - ¿Qué tecnología se usará?
      - ¿Qué restricciones técnicas existen?
      - ¿Qué debe evitar la AI?
```

---

## 29. División ideal para desarrollarlo con AI

### Prompt 1 — Crear proyecto base

```txt
Crea una CLI en Go llamada apd usando Cobra.
Debe tener comandos new, resume, export, backlog y prompts.
Por ahora los comandos pueden imprimir mensajes placeholder.
Incluye estructura limpia de carpetas.
```

### Prompt 2 — Cargar plantillas YAML

```txt
Implementa un loader de plantillas YAML.
Debe leer archivos desde /templates.
Cada plantilla tiene id, name, description y sections.
Valida que cada sección tenga id y title.
```

### Prompt 3 — Motor de preguntas

```txt
Implementa un motor que recorra las secciones de una plantilla.
Debe mostrar title, description, help, example y questions.
Debe permitir escribir respuesta multilinea.
Debe soportar /skip, /help, /back y /done.
```

### Prompt 4 — Guardado de sesión

```txt
Implementa guardado automático de sesión en .apd/sessions.
Cada respuesta debe persistirse para poder continuar luego con apd resume.
```

### Prompt 5 — Generador Markdown

```txt
Implementa un exportador Markdown que tome un Document y genere un archivo .md limpio.
Debe incluir metadata, secciones respondidas, secciones saltadas opcionalmente y un AI Context Pack.
```

### Prompt 6 — Backlog básico

```txt
Implementa un generador inicial de backlog basado en capacidades y reglas.
Debe producir épicas, historias de usuario, criterios de aceptación y tareas técnicas sugeridas.
No debe inventar demasiado: si falta información, debe marcarlo como pendiente.
```

### Prompt 7 — Prompt Pack

```txt
Implementa generación de prompts para AI a partir del documento.
Debe generar prompts para arquitectura, entidades, API, UI, tests y revisión.
```

---

## 30. Decisión importante

No se recomienda venderlo ni pensarlo como:

```txt
PRD Generator
```

Eso lo limita.

Mejor como:

```txt
AI Product Decomposer
```

O:

```txt
Structured Product Mapper
```

Porque sirve para:

- PRD
- CR
- feature spec
- backlog
- prompts
- documentación técnica
- análisis funcional
- diseño inicial
- refactor
- bugs
- módulos

---

## 31. Resumen ejecutivo

La herramienta debe resolver esto:

```txt
Convertir ideas, cambios o problemas en documentación estructurada, flexible y útil para construir software con AI.
```

Su valor principal no es generar Markdown.

Su valor principal es guiar el pensamiento del usuario.

El Markdown es solo el resultado.

El producto real es el flujo:

```txt
Pensar → ordenar → descomponer → documentar → generar contexto → construir
```
