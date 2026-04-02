# Nano-Paas
# 🌌 Nebula: Nano-PaaS
> **Extreme Lightweight Self-Hosted PaaS** > *Despliegue de aplicaciones profesional para hardware humilde.*

**Autor:** Mohamed Kamil El Kouarti Mechhidan  
**Grado:** 2º SMR - Prometeo (2025-2026)  
**Proyecto:** Sistemas Microinformáticos y Redes

---

## 🚀 El Concepto
**Nebula** es una plataforma de "Plataforma como Servicio" (PaaS) diseñada para competir en un nicho que los gigantes como Heroku o Coolify han olvidado: **la eficiencia extrema de recursos.**

Mientras que otras soluciones requieren 2GB+ de RAM solo para el panel de control, Nebula está optimizado para ejecutarse en **Raspberry Pi, VPS económicos (512MB RAM) o hardware reciclado**, permitiendo desplegar aplicaciones web con un consumo mínimo.
Una plataforma de servicios en la nube (PaaS) ultra-ligera diseñada para transformar hardware modesto en una infraestructura empresarial segura, privada y fácil de gestionar.


## ✨ Características Principales
* **Nano-Footprint:** El núcleo del sistema consume menos de 20MB de RAM.
* **Nebula Pulse (Go):** Monitor de sistema propietario escrito en Go para máxima eficiencia.
* **Security-First:** Integración nativa con `fail2ban`, cortafuegos pre-configurado y gestión automática de SSL vía Nginx Proxy Manager.
* **Multi-Stack:** Soporte para contenedores Docker (WordPress, Python, Static Web, etc.).
* **Zero-Conf CI/CD:** Flujo de despliegue automatizado basado en Git.

## 🛠️ Stack Tecnológico
Para lograr esta ligereza, Nebula selecciona cuidadosamente sus componentes:
* **Motor:** Docker & Docker Compose.
* **Proxy/Ingress:** Nginx Proxy Manager (Gestión de tráfico y certificados).
* **Core Monitor (Pulse):** Golang (Sustituyendo soluciones pesadas como NetData).
* **Seguridad:** Fail2Ban & Scripts de Hardening.

---

## 📉 El Diferenciador: Go vs. El Resto
Originalmente, el monitoreo se planteó con herramientas estándar de la industria, pero estas consumían hasta 500MB de RAM. 

**Nebula Pulse** ha sido re-escrito en **Go** para:
1.  **Reducir el consumo de RAM en un 95%.**
2.  **Eliminar dependencias:** Se distribuye como un único binario estático.
3.  **Velocidad nativa:** Respuesta instantánea en la telemetría del sistema.

---

## 🏗️ Estructura del Proyecto
```bash
.
├── apps/                # Plantillas de aplicaciones optimizadas
├── pulse/               # Monitor de recursos escrito en Go
├── config/              # Configuraciones de seguridad (Fail2Ban, Nginx)
├── scripts/             # Automatización de despliegue y backups
└── docker-compose.yml   # Orquestación del núcleo de Nebula
