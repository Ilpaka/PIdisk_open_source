<h1 align="center">PIdisk</h1>

<p align="center">
  <a href="README.md"><img src="https://img.shields.io/badge/lang-English-555?style=for-the-badge" alt="English"/></a>
  <a href="README.ru.md"><img src="https://img.shields.io/badge/язык-Русский-1c74d4?style=for-the-badge" alt="Русский"/></a>
</p>

<p align="center">
  <b>Кросс-платформенный SFTP-менеджер, который не мешает работать.</b><br/>
  Двусторонняя синхронизация папок, аутентификация только по ключу, секреты в OS keyring. macOS, Windows, Linux.
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white">
  <img alt="React" src="https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white">
  <img alt="MUI" src="https://img.shields.io/badge/MUI-6-007FFF?logo=mui&logoColor=white">
  <img alt="Wails" src="https://img.shields.io/badge/Wails-v2-DF0000?logo=go&logoColor=white">
  <img alt="License" src="https://img.shields.io/badge/license-MIT-green">
  <img alt="Platforms" src="https://img.shields.io/badge/platforms-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey">
</p>

<p align="center">
  <img alt="Файловый менеджер PIdisk" src="docs/screenshots/files-dark.png" width="820"/>
</p>

---

## Что это

PIdisk подключается к серверу по SSH, даёт работать с удалёнными файлами как с обычным файловым менеджером и синхронизирует выбранные папки между твоей машиной и сервером. Без веб-интерфейса, без Docker, без агентов на сервере. Только SSH.

## Возможности

- **SSH только по ключу** с TOFU-проверкой fingerprint. Для каждого профиля автоматически генерируется новый Ed25519-ключ.
- **Двусторонняя синхронизация** с разрешением конфликтов по последнему изменению (LWW) и `.pidiskignore` (синтаксис gitignore).
- **Drag and drop** файлов прямо в дерево папок.
- **Корзина с восстановлением.** Удалённое попадает в персональную корзину профиля и возвращается на прежнее место одной кнопкой.
- **Параллельные передачи**, утилизируют гигабит. Прогресс-бар, отмена на лету.
- **Тёмная тема** и хоткеи: F2 переименовать, Del в корзину, Esc снять выделение, Ctrl/Cmd+A выделить всё, F5 обновить.
- **Регулируемая ширина дерева** папок с сохранением между запусками.
- **Авто-переподключение** при разрывах сети без потери активного профиля.
- **OS keyring** (macOS Keychain / Credential Manager / Secret Service) для всех passphrase.
- **Нативные сборки** под macOS (.app), Windows (.exe), Linux (AppImage / бинарь).

## Скриншоты

<table>
  <tr>
    <td align="center">
      <img alt="Логин" src="docs/screenshots/login.png" width="380"/><br/>
      <sub>Логин: выбор существующего профиля или создание нового</sub>
    </td>
    <td align="center">
      <img alt="Создание профиля" src="docs/screenshots/create-profile.png" width="380"/><br/>
      <sub>Создание профиля: четыре поля, остальное подставляется</sub>
    </td>
  </tr>
  <tr>
    <td align="center">
      <img alt="Файловый менеджер" src="docs/screenshots/files-dark.png" width="380"/><br/>
      <sub>Файлы: дерево с ресайзом, breadcrumbs, drag and drop</sub>
    </td>
    <td align="center">
      <img alt="Передачи" src="docs/screenshots/transfers.png" width="380"/><br/>
      <sub>Передачи: живой прогресс, отмена в любой момент</sub>
    </td>
  </tr>
</table>

## Установка

Нативные сборки готовит GitHub Actions на каждый тег. Скачать можно
со страницы [Releases](../../releases).

## Сборка из исходников

```bash
# Toolchain
brew install go node                                                  # macOS
go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0

# Только Linux
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# Запуск с hot reload
cd frontend && npm install && cd ..
wails dev

# Релизная сборка под текущую платформу
wails build -clean
```

## Как это устроено

```mermaid
flowchart LR
    UI[React + MUI<br/>фронтенд] -- Wails IPC --> Backend[Go-бэкенд]
    Backend -- SSH/SFTP --> Server[Удалённый сервер]
    Backend -. keyring .-> OS[OS Keyring]
    Backend -. метаданные .-> Local[(bbolt-файлы)]
```

Один Go-бинарь со встроенным React-фронтом ходит на сервер через
`golang.org/x/crypto/ssh` плюс `pkg/sftp`. Секреты лежат в нативном
OS keyring; профили и метаданные корзины — в нескольких маленьких
bbolt-файлах в пользовательской директории данных.

## Документация

- [Архитектура](docs/ARCHITECTURE.md): структура, направление зависимостей, поток событий.
- [Безопасность](docs/SECURITY.md): модель угроз и что намеренно не поддерживается.
- [Профили](docs/PROFILES.md): жизненный цикл, формат хранения, работа с keyring.
- [Синхронизация](docs/SYNC.md): цикл, алгоритм diff, обработка ignore.

## Roadmap

- Встроенная вкладка терминала (xterm.js)
- Закладки на часто используемые пути
- Двухпанельный режим (side by side)
- Версионирование удалений в корзине
- Импорт / экспорт профилей

## Лицензия

MIT. См. [LICENSE](./LICENSE).
