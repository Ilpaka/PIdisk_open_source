// src-tauri/src/main.rs

#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use once_cell::sync::Lazy;
use serde::{Deserialize, Serialize};
use ssh2::Session;
use std::io::Read;
use std::net::TcpStream;
use std::sync::Mutex;
use tauri::command;
use std::fs::File;
use std::io::{Write, BufWriter};
use std::fs;
use std::io::BufReader;
use serde_json;
use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::time::Duration;
use std::sync::Arc;
use tokio::sync::RwLock;
use notify::{Watcher, RecursiveMode, RecommendedWatcher, Event};
use std::path::{Path, PathBuf};
use std::env;
use tokio::sync::mpsc;
use walkdir::WalkDir;

/// SSH-настройки + корень и корзина + синхронизация
#[derive(Serialize, Deserialize, Clone)]
struct Settings {
    host: String,
    port: u16,
    username: String,
    password: String,
    root_dir: String,
    trash_dir: String,
    sync_enabled: bool,
    sync_interval_seconds: u64,
    local_sync_dir: String, // Локальная папка Pidisk
}

/// Информация о синхронизируемой папке
#[derive(Serialize, Deserialize, Clone, Debug)]
struct SyncFolder {
    name: String,
    local_path: String,
    remote_path: String,
    enabled: bool,
    last_sync: u64, // timestamp
}

/// Информация о файле для синхронизации
#[derive(Serialize, Deserialize, Clone, Debug)]
struct FileInfo {
    path: String,
    modified: u64,
    size: u64,
    is_directory: bool,
}

/// Статус синхронизации
#[derive(Serialize, Deserialize, Clone)]
struct SyncStatus {
    is_running: bool,
    last_sync_time: u64,
    synced_folders: Vec<String>,
    errors: Vec<String>,
    files_synced: u32,
    bytes_synced: u64,
}

// Функция для определения пути к settings.json
fn get_settings_path() -> String {
    // Пробуем несколько возможных путей
    let possible_paths = vec![
        "src-tauri/settings.json",  // Относительно корня проекта
        "settings.json",             // В текущей директории
        "../settings.json",          // На уровень выше
    ];
    
    // Сначала пробуем относительные пути
    for path_str in &possible_paths {
        if Path::new(path_str).exists() {
            return path_str.to_string();
        }
    }
    
    // Если не нашли, используем путь относительно текущей директории
    if let Ok(current_dir) = env::current_dir() {
        let path = current_dir.join("src-tauri").join("settings.json");
        if path.exists() {
            return path.to_string_lossy().to_string();
        }
        let path = current_dir.join("settings.json");
        if path.exists() {
            return path.to_string_lossy().to_string();
        }
    }
    
    // По умолчанию используем src-tauri/settings.json
    "src-tauri/settings.json".to_string()
}

// Глобальные переменные с исправленным доступом к RwLock
static SETTINGS: Lazy<Arc<RwLock<Settings>>> = Lazy::new(|| {
    let path = get_settings_path();
    println!("Загрузка настроек из: {}", path);
    let settings = if Path::new(&path).exists() {
        load_settings_from_file(&path)
    } else {
        println!("Файл настроек не найден, используем настройки по умолчанию");
        let default = default_settings();
        // Сохраняем настройки по умолчанию в найденный путь
        if let Some(parent) = Path::new(&path).parent() {
            let _ = fs::create_dir_all(parent);
        }
        save_settings_to_file(&path, &default);
        default
    };
    Arc::new(RwLock::new(settings))
});

static SSH_SESSION: Lazy<Arc<Mutex<Option<Session>>>> = Lazy::new(|| Arc::new(Mutex::new(None)));
static CURRENT_DIR: Lazy<Arc<RwLock<String>>> = Lazy::new(|| {
    Arc::new(RwLock::new("/root/PIdisk".to_string()))
});

// Синхронизация
static SYNC_FOLDERS: Lazy<Arc<RwLock<HashMap<String, SyncFolder>>>> = Lazy::new(|| {
    Arc::new(RwLock::new(HashMap::new()))
});

static SYNC_STATUS: Lazy<Arc<RwLock<SyncStatus>>> = Lazy::new(|| {
    Arc::new(RwLock::new(SyncStatus {
        is_running: false,
        last_sync_time: 0,
        synced_folders: Vec::new(),
        errors: Vec::new(),
        files_synced: 0,
        bytes_synced: 0,
    }))
});

static SYNC_SHUTDOWN: Lazy<Arc<RwLock<bool>>> = Lazy::new(|| Arc::new(RwLock::new(false)));

fn load_settings_from_file(path: &str) -> Settings {
    match fs::read_to_string(path) {
        Ok(data) => serde_json::from_str(&data).unwrap_or_else(|_| default_settings()),
        Err(_) => default_settings(),
    }
}

fn save_settings_to_file(path: &str, settings: &Settings) {
    let data = serde_json::to_string_pretty(settings).unwrap();
    fs::write(path, data).unwrap();
}

// fn default_settings() -> Settings {
//     let path = "/Users/sip/Yandex.Disk.localized/Learning/Projects/PIdisk/pidisk-app/src-tauri/src/settings.json";
//     let file = File::open(path).expect("Не удалось открыть settings.json");
//     let reader = BufReader::new(file);
//     let mut settings: Settings = serde_json::from_reader(reader).expect("Ошибка парсинга settings.json");
    
//     settings.sync_enabled = false;
//     settings.sync_interval_seconds = 30;
//     settings.local_sync_dir = "/Users/sip/Pidisk".to_string();
    
//     settings
// }


fn default_settings() -> Settings {
    // Определяем домашнюю директорию пользователя динамически
    let local_sync_dir = dirs::home_dir()
        .map(|p| p.join("Pidisk").to_string_lossy().to_string())
        .unwrap_or_else(|| format!("{}/Pidisk", env::var("HOME").unwrap_or_else(|_| "/tmp".to_string())));
    
    Settings {
        host: "138.124.14.1".to_string(),
        port: 22,
        username: "root".to_string(),
        password: "UHxeVf4KXDHz".to_string(),
        root_dir: "/root/PIdisk".to_string(),
        trash_dir: "/root/PIdisk/Bin".to_string(),
        sync_enabled: true,
        sync_interval_seconds: 30,
        local_sync_dir,
    }
}

/// Создаёт новую SSH-сессию
fn create_session(cfg: &Settings) -> Result<Session, String> {
    let addr = format!("{}:{}", cfg.host, cfg.port);
    println!("Подключение к адресу: {}:{}", cfg.host, cfg.port);
    let tcp = TcpStream::connect(&addr).map_err(|e| e.to_string())?;
    let mut sess = Session::new().map_err(|e| e.to_string())?;
    sess.set_tcp_stream(tcp);
    sess.handshake().map_err(|e| e.to_string())?;
    sess.userauth_password(&cfg.username, &cfg.password)
        .map_err(|e| e.to_string())?;
    Ok(sess)
}

/// Утилита для работы с SSH-сессией - исправлена ошибка borrowed data escapes
async fn with_session<F, R>(f: F) -> Result<R, String>
where
    F: FnOnce(&mut Session) -> Result<R, String> + Send + 'static,
    R: Send + 'static,
{
    let settings = SETTINGS.read().await.clone();
    
    tokio::task::spawn_blocking(move || {
        let mut guard = SSH_SESSION.lock().unwrap();
        if guard.is_none() {
            *guard = Some(create_session(&settings)?);
        }
        let sess = guard.as_mut().unwrap();
        f(sess)
    }).await.map_err(|e| e.to_string())?
}

// ----------------------
// Команды синхронизации
// ----------------------

/// Добавить папку для синхронизации
#[command]
async fn add_sync_folder(name: String, local_path: String, remote_path: String) -> Result<(), String> {
    println!("📁 Добавляем папку для синхронизации: {} -> {}", local_path, remote_path);
    
    let sync_folder = SyncFolder {
        name: name.clone(),
        local_path: local_path.clone(),
        remote_path,
        enabled: true,
        last_sync: 0,
    };
    
    let mut folders = SYNC_FOLDERS.write().await;
    folders.insert(name.clone(), sync_folder);
    println!("📁 Папка {} добавлена в список", name);
    
    // Создаем локальную папку если не существует
    tokio::fs::create_dir_all(&local_path).await
        .map_err(|e| format!("Не удалось создать локальную папку: {}", e))?;
    
    println!("✅ Локальная папка создана: {}", local_path);
    Ok(())
}


/// Получить список папок для синхронизации
#[command]
async fn get_sync_folders() -> Vec<SyncFolder> {
    SYNC_FOLDERS.read().await.values().cloned().collect()
}

/// Включить/выключить синхронизацию папки
#[command]
async fn toggle_sync_folder(name: String, enabled: bool) -> Result<(), String> {
    let mut folders = SYNC_FOLDERS.write().await;
    if let Some(folder) = folders.get_mut(&name) {
        folder.enabled = enabled;
    }
    Ok(())
}

/// Удалить папку из синхронизации
#[command]
async fn remove_sync_folder(name: String) -> Result<(), String> {
    let mut folders = SYNC_FOLDERS.write().await;
    folders.remove(&name);
    Ok(())
}

/// Получить статус синхронизации
#[command]
async fn get_sync_status() -> SyncStatus {
    SYNC_STATUS.read().await.clone()
}

/// Запустить синхронизацию
#[command]
async fn start_sync() -> Result<(), String> {
    println!("🔄 Команда start_sync вызвана");
    
    let mut status = SYNC_STATUS.write().await;
    if status.is_running {
        println!("❌ Синхронизация уже запущена");
        return Err("Синхронизация уже запущена".to_string());
    }
    
    status.is_running = true;
    status.errors.clear();
    drop(status);
    
    *SYNC_SHUTDOWN.write().await = false;
    
    let settings = SETTINGS.read().await.clone();
    println!("📋 Настройки загружены: sync_enabled={}", settings.sync_enabled);
    
    let folders = SYNC_FOLDERS.read().await.clone();
    println!("📁 Папок для синхронизации: {}", folders.len());
    
    if folders.is_empty() {
        println!("⚠️ Нет папок для синхронизации!");
        return Err("Сначала добавьте папки для синхронизации".to_string());
    }
    
    tokio::spawn(async move {
        println!("🚀 Запускаем цикл синхронизации");
        sync_loop(settings).await;
    });
    
    setup_file_watcher().await?;
    
    println!("✅ Синхронизация запущена успешно!");
    Ok(())
}



/// Остановить синхронизацию
#[command]
async fn stop_sync() -> Result<(), String> {
    *SYNC_SHUTDOWN.write().await = true;
    let mut status = SYNC_STATUS.write().await;
    status.is_running = false;
    Ok(())
}

/// Основной цикл синхронизации
async fn sync_loop(settings: Settings) {
    println!("🔄 Цикл синхронизации запущен");
    let interval = Duration::from_secs(settings.sync_interval_seconds);
    
    while !*SYNC_SHUTDOWN.read().await {
        println!("🔄 Выполняем итерацию синхронизации");
        let folders = SYNC_FOLDERS.read().await.clone();
        let mut files_synced = 0u32;
        let mut bytes_synced = 0u64;
        
        for (name, folder) in folders {
            if !folder.enabled {
                println!("⏭️ Пропускаем отключенную папку: {}", name);
                continue;
            }
            
            println!("🔄 Синхронизируем папку: {}", name);
            match sync_folder(&folder).await {
                Ok((files, bytes)) => {
                    files_synced += files;
                    bytes_synced += bytes;
                    println!("✅ Папка {} синхронизирована: {} файлов, {} байт", name, files, bytes);
                }
                Err(e) => {
                    println!("❌ Ошибка синхронизации {}: {}", name, e);
                    let mut status = SYNC_STATUS.write().await;
                    status.errors.push(format!("Ошибка синхронизации {}: {}", name, e));
                }
            }
        }
        
        // Обновляем статус синхронизации
        let mut status = SYNC_STATUS.write().await;
        status.last_sync_time = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();
        status.files_synced = files_synced;
        status.bytes_synced = bytes_synced;
        
        println!("📊 Статистика синхронизации: {} файлов, {} байт", files_synced, bytes_synced);
        
        tokio::time::sleep(interval).await;
    }
    
    println!("🛑 Цикл синхронизации остановлен");
    let mut status = SYNC_STATUS.write().await;
    status.is_running = false;
}

/// Синхронизация одной папки
async fn sync_folder(folder: &SyncFolder) -> Result<(u32, u64), String> {
    let mut files_synced = 0u32;
    let mut bytes_synced = 0u64;
    
    // 1. Синхронизация локальные изменения -> сервер
    let (local_files, local_bytes) = sync_local_to_remote(folder).await?;
    files_synced += local_files;
    bytes_synced += local_bytes;
    
    // 2. Синхронизация сервер -> локальные изменения  
    let (remote_files, remote_bytes) = sync_remote_to_local(folder).await?;
    files_synced += remote_files;
    bytes_synced += remote_bytes;
    
    Ok((files_synced, bytes_synced))
}

/// Получить информацию о файлах в локальной папке
async fn get_local_files(path: &str) -> Result<HashMap<String, FileInfo>, String> {
    let mut files = HashMap::new();
    
    for entry in WalkDir::new(path).into_iter().filter_map(|e| e.ok()) {
        let file_path = entry.path();
        let relative_path = file_path.strip_prefix(path)
            .map_err(|e| e.to_string())?
            .to_string_lossy()
            .to_string();
        
        if relative_path.is_empty() {
            continue;
        }
        
        let metadata = tokio::fs::metadata(file_path).await
            .map_err(|e| e.to_string())?;
        
        let modified = metadata.modified()
            .map_err(|e| e.to_string())?
            .duration_since(UNIX_EPOCH)
            .map_err(|e| e.to_string())?
            .as_secs();
        
        files.insert(relative_path, FileInfo {
            path: file_path.to_string_lossy().to_string(),
            modified,
            size: metadata.len(),
            is_directory: metadata.is_dir(),
        });
    }
    
    Ok(files)
}

/// Получить информацию о файлах на сервере - исправлена ошибка borrowed data escapes
async fn get_remote_files(remote_path: &str) -> Result<HashMap<String, FileInfo>, String> {
    let remote_path_owned = remote_path.to_string();
    
    with_session(move |sess| {
        let mut files = HashMap::new();
        
        // Используем find для получения всех файлов с метаданными
        let cmd = format!("find '{}' -type f -exec stat -c '%n|%Y|%s' {{}} \\;", remote_path_owned);
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&cmd).map_err(|e| e.to_string())?;
        
        let mut out = String::new();
        ch.read_to_string(&mut out).map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;
        
        for line in out.lines() {
            let parts: Vec<&str> = line.split('|').collect();
            if parts.len() == 3 {
                let full_path = parts[0];
                let modified: u64 = parts[1].parse().unwrap_or(0);
                let size: u64 = parts[2].parse().unwrap_or(0);
                
                if let Some(relative_path) = full_path.strip_prefix(&format!("{}/", remote_path_owned)) {
                    files.insert(relative_path.to_string(), FileInfo {
                        path: full_path.to_string(),
                        modified,
                        size,
                        is_directory: false,
                    });
                }
            }
        }
        
        Ok(files)
    }).await
}

/// Синхронизация локальных изменений на сервер
async fn sync_local_to_remote(folder: &SyncFolder) -> Result<(u32, u64), String> {
    let local_files = get_local_files(&folder.local_path).await?;
    let remote_files = get_remote_files(&folder.remote_path).await?;
    
    let mut files_synced = 0u32;
    let mut bytes_synced = 0u64;
    
    for (relative_path, local_file) in local_files {
        if local_file.is_directory {
            continue;
        }
        
        let should_upload = match remote_files.get(&relative_path) {
            Some(remote_file) => local_file.modified > remote_file.modified, // Конфликт: берем последний измененный
            None => true, // Файл не существует на сервере
        };
        
        if should_upload {
            let file_data = tokio::fs::read(&local_file.path).await
                .map_err(|e| format!("Не удалось прочитать файл: {}", e))?;
            
            upload_file_to_remote(&relative_path, file_data, &folder.remote_path).await?;
            files_synced += 1;
            bytes_synced += local_file.size;
        }
    }
    
    Ok((files_synced, bytes_synced))
}

/// Синхронизация изменений с сервера
async fn sync_remote_to_local(folder: &SyncFolder) -> Result<(u32, u64), String> {
    let local_files = get_local_files(&folder.local_path).await?;
    let remote_files = get_remote_files(&folder.remote_path).await?;
    
    let mut files_synced = 0u32;
    let mut bytes_synced = 0u64;
    
    for (relative_path, remote_file) in remote_files {
        let should_download = match local_files.get(&relative_path) {
            Some(local_file) => remote_file.modified > local_file.modified, // Конфликт: берем последний измененный
            None => true, // Файл не существует локально
        };
        
        if should_download {
            let local_file_path = Path::new(&folder.local_path).join(&relative_path);
            
            // Создаем директории если нужно
            if let Some(parent) = local_file_path.parent() {
                tokio::fs::create_dir_all(parent).await
                    .map_err(|e| format!("Не удалось создать директорию: {}", e))?;
            }
            
            download_file_from_remote(&relative_path, &folder.remote_path, &local_file_path).await?;
            files_synced += 1;
            bytes_synced += remote_file.size;
        }
    }
    
    Ok((files_synced, bytes_synced))
}

/// Загрузка файла на сервер
async fn upload_file_to_remote(filename: &str, data: Vec<u8>, remote_dir: &str) -> Result<(), String> {
    let filename = filename.to_string();
    let remote_dir = remote_dir.to_string();
    
    with_session(move |sess| {
        let remote_path = format!("{}/{}", remote_dir, filename);
        
        // Создаем директории на сервере если нужно
        if let Some(parent) = Path::new(&remote_path).parent() {
            let mkdir_cmd = format!("mkdir -p '{}'", parent.to_string_lossy());
            let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
            ch.exec(&mkdir_cmd).map_err(|e| e.to_string())?;
            ch.close().map_err(|e| e.to_string())?;
            ch.wait_close().map_err(|e| e.to_string())?;
        }
        
        let size = data.len() as u64;
        let mut remote = sess
            .scp_send(Path::new(&remote_path), 0o644, size, None)
            .map_err(|e| e.to_string())?;
        std::io::Write::write_all(&mut remote, &data)
            .map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

/// Скачивание файла с сервера
async fn download_file_from_remote(filename: &str, remote_dir: &str, local_path: &Path) -> Result<(), String> {
    let remote_file_path = format!("{}/{}", remote_dir, filename);
    let local_path_str = local_path.to_string_lossy().to_string();
    
    with_session(move |sess| {
        let sftp = sess.sftp().map_err(|e| e.to_string())?;
        let mut remote = sftp.open(&remote_file_path).map_err(|e| e.to_string())?;
        
        let file = std::fs::File::create(&local_path_str).map_err(|e| e.to_string())?;
        let mut writer = BufWriter::new(file);
        
        let mut buf = [0u8; 16 * 1024];
        loop {
            let n = remote.read(&mut buf).map_err(|e| e.to_string())?;
            if n == 0 { break; }
            writer.write_all(&buf[..n]).map_err(|e| e.to_string())?;
        }
        
        Ok(())
    }).await
}

/// Настройка файлового наблюдателя
async fn setup_file_watcher() -> Result<(), String> {
    let folders = SYNC_FOLDERS.read().await.clone();
    
    for (_, folder) in folders {
        if !folder.enabled {
            continue;
        }
        
        let folder_clone = folder.clone();
        tokio::spawn(async move {
            if let Err(e) = watch_folder(folder_clone).await {
                println!("Ошибка файлового наблюдателя: {}", e);
            }
        });
    }
    
    Ok(())
}

/// Наблюдение за изменениями в папке - исправлена ошибка с unwrap_or
async fn watch_folder(folder: SyncFolder) -> Result<(), String> {
    let (tx, mut rx) = mpsc::channel(100);
    
    let folder_path = folder.local_path.clone();
    let watcher_tx = tx.clone();
    
    tokio::task::spawn_blocking(move || {
        let mut watcher = RecommendedWatcher::new(
            move |res: Result<Event, notify::Error>| {
                if let Err(_) = watcher_tx.blocking_send(res) {
                    // Канал закрыт, выходим
                }
            },
            notify::Config::default(),
        ).map_err(|e| e.to_string())?;
        
        watcher.watch(Path::new(&folder_path), RecursiveMode::Recursive)
            .map_err(|e| e.to_string())?;
        
        // Держим watcher живым
        loop {
            std::thread::sleep(std::time::Duration::from_secs(1));
            // Исправлена ошибка с unwrap_or - используем правильный тип
            if SYNC_SHUTDOWN.try_read().map(|guard| *guard).unwrap_or(true) {
                break;
            }
        }
        
        Ok::<(), String>(())
    });
    
    while let Some(event_result) = rx.recv().await {
        if *SYNC_SHUTDOWN.read().await {
            break;
        }
        
        match event_result {
            Ok(event) => {
                // Обрабатываем только события создания и изменения файлов
                if event.kind.is_create() || event.kind.is_modify() {
                    for path in event.paths {
                        if path.is_file() {
                            if let Some(file_name) = path.file_name() {
                                if let Some(file_name_str) = file_name.to_str() {
                                    // Небольшая задержка чтобы файл успел записаться
                                    tokio::time::sleep(Duration::from_millis(500)).await;
                                    
                                    if let Ok(data) = tokio::fs::read(&path).await {
                                        let relative_path = path.strip_prefix(&folder.local_path)
                                            .map(|p| p.to_string_lossy().to_string())
                                            .unwrap_or_else(|_| file_name_str.to_string());
                                        
                                        let _ = upload_file_to_remote(&relative_path, data, &folder.remote_path).await;
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Err(_) => {}
        }
    }
    
    Ok(())
}

// Остальные команды - исправлены ошибки с доступом к RwLock
#[command]
async fn get_settings() -> Settings {
    SETTINGS.read().await.clone()
}

#[command]
async fn update_settings(
    host: String,
    port: u16,
    username: String,
    password: String,
) -> Result<(), String> {
    let mut trial = SETTINGS.read().await.clone();
    trial.host = host.clone();
    trial.port = port;
    trial.username = username.clone();
    trial.password = password.clone();

    if trial.host.is_empty() || trial.port == 0 || trial.username.is_empty() {
        return Err("Заполните все обязательные поля в settings.json".into());
    }

    let sess = create_session(&trial)
        .map_err(|e| format!("Неверные данные для подключения: {}", e))?;

    let mut cfg = SETTINGS.write().await;
    cfg.host = host;
    cfg.port = port;
    cfg.username = username;
    cfg.password = password;

    *SSH_SESSION.lock().unwrap() = Some(sess);

    let root = cfg.root_dir.clone();
    *CURRENT_DIR.write().await = root;

    Ok(())
}

#[command]
async fn read_dir(dir: String) -> Result<(String, Vec<String>), String> {
    let result = with_session(move |sess| {
        let cmd = format!("cd '{}' && pwd && ls -1", dir);
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&cmd).map_err(|e| e.to_string())?;
        let mut out = String::new();
        ch.read_to_string(&mut out).map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;

        let mut lines = out.lines();
        let new_path = lines.next().unwrap_or(&dir).to_string();
        let list = lines.map(String::from).collect();

        Ok((new_path, list))
    }).await?;
    
    // Обновляем текущую директорию
    *CURRENT_DIR.write().await = result.0.clone();
    
    Ok(result)
}

/// mkdir в текущей папке - исправлена ошибка с async
#[command]
async fn mkdir(name: String) -> Result<(), String> {
    let cwd = CURRENT_DIR.read().await.clone();
    
    with_session(move |sess| {
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&format!("cd '{}' && mkdir '{}'", cwd, name))
            .map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

/// mv (переименовать или переместить) - исправлена ошибка с async
#[command]
async fn mv(src: String, dest: String) -> Result<(), String> {
    let cwd = CURRENT_DIR.read().await.clone();
    
    with_session(move |sess| {
        let src_full = if src.starts_with('/') {
            src
        } else {
            format!("{}/{}", cwd, src)
        };
        
        let dest_full = if dest.starts_with('/') {
            dest
        } else {
            format!("{}/{}", cwd, dest)
        };
        
        let cmd = format!("sh -lc \"mv '{}' '{}'\"", src_full, dest_full);
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&cmd).map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

/// rm: в trash_dir или удаление внутри корзины - исправлена ошибка с async
#[command]
async fn rm(target: String) -> Result<(), String> {
    let trash = SETTINGS.read().await.trash_dir.clone();
    let cwd = CURRENT_DIR.read().await.clone();
    
    with_session(move |sess| {
        // создаём корзину при необходимости
        let mut c0 = sess.channel_session().map_err(|e| e.to_string())?;
        c0.exec(&format!("mkdir -p '{}'", trash))
            .map_err(|e| e.to_string())?;
        c0.close().map_err(|e| e.to_string())?;
        c0.wait_close().map_err(|e| e.to_string())?;

        let cmd = if cwd == trash {
            format!("sh -lc \"cd '{}' && rm -rf '{}'\"", trash, target)
        } else {
            let src_full = format!("{}/{}", cwd, target);
            format!("sh -lc \"cd '{}' && mv '{}' '{}'\"", cwd, src_full, trash)
        };

        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&cmd).map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

/// clear_all внутри trash_dir - исправлена ошибка с async
#[command]
async fn clear_all() -> Result<(), String> {
    let trash = SETTINGS.read().await.trash_dir.clone();
    
    with_session(move |sess| {
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        let cmd = format!("cd '{}' && rm -rf ./*", trash);
        ch.exec(&cmd).map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

/// df -h текущей папки - исправлена ошибка с async
#[command]
async fn df() -> Result<String, String> {
    let cwd = CURRENT_DIR.read().await.clone();
    
    with_session(move |sess| {
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&format!("cd '{}' && df -h .", cwd))
            .map_err(|e| e.to_string())?;
        let mut out = String::new();
        ch.read_to_string(&mut out).map_err(|e| e.to_string())?;
        Ok(out)
    }).await
}

#[command]
async fn upload_file(filename: String, data: Vec<u8>) -> Result<(), String> {
    let cwd = CURRENT_DIR.read().await.clone();
    let remote_path = std::path::Path::new(&cwd).join(&filename);
    
    with_session(move |sess| {
        let size = data.len() as u64;
        let mut remote = sess
            .scp_send(&remote_path, 0o644, size, None)
            .map_err(|e| e.to_string())?;
        std::io::Write::write_all(&mut remote, &data)
            .map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

#[command]
async fn download_and_save(server_file_name: String, save_path: String) -> Result<(), String> {
    let base_dir = CURRENT_DIR.read().await.clone();
    let full_server_path = std::path::Path::new(&base_dir).join(&server_file_name);
    
    let file = File::create(&save_path).map_err(|e| e.to_string())?;
    let mut writer = BufWriter::new(file);
    
    with_session(move |sess| {
        let sftp = sess.sftp().map_err(|e| e.to_string())?;
        let mut remote = sftp.open(full_server_path.to_str().unwrap()).map_err(|e| {
            let err_msg = format!("Failed to open remote file '{:?}': {}", full_server_path, e);
            println!("{}", err_msg);
            err_msg
        })?;
        
        let mut buf = [0u8; 16 * 1024];
        loop {
            let n = remote.read(&mut buf).map_err(|e| e.to_string())?;
            if n == 0 { break; }
            writer.write_all(&buf[..n]).map_err(|e| e.to_string())?;
        }
        
        Ok(())
    }).await
}

#[tauri::command]
async fn download_folder(server_folder_name: String, save_path: String) -> Result<(), String> {
    let settings = SETTINGS.read().await.clone();
    let current_dir = CURRENT_DIR.read().await.clone();
    
    // Формируем путь к папке на сервере
    let remote_folder_path = std::path::Path::new(&current_dir).join(&server_folder_name);
    let remote_folder_str = remote_folder_path.to_str().ok_or("Некорректный путь к папке на сервере")?;
    
    // Формируем адрес для scp: user@host:/remote/path
    let remote = format!("{}@{}:{}", settings.username, settings.host, remote_folder_str);
    
    // Запускаем команду scp -P порт -r user@host:/remote/path /local/path
    let status = std::process::Command::new("scp")
        .arg("-P").arg(settings.port.to_string())
        .arg("-r")
        .arg(&remote)
        .arg(&save_path)
        .status()
        .map_err(|e| format!("Не удалось запустить scp: {}", e))?;
    
    if status.success() {
        Ok(())
    } else {
        Err(format!("scp завершился с ошибкой: {}", status))
    }
}

#[command]
async fn rename(old: String, new: String) -> Result<(), String> {
    let cwd = CURRENT_DIR.read().await.clone();
    let cmd = format!("sh -lc \"cd '{}' && mv '{}' '{}'\"", cwd, old, new);
    
    with_session(move |sess| {
        let mut ch = sess.channel_session().map_err(|e| e.to_string())?;
        ch.exec(&cmd).map_err(|e| e.to_string())?;
        ch.close().map_err(|e| e.to_string())?;
        ch.wait_close().map_err(|e| e.to_string())?;
        Ok(())
    }).await
}

#[tokio::main]
async fn main() {
    let path = get_settings_path();
    println!("Используемый settings.json: {}", path);
    let file_content = std::fs::read_to_string(&path).unwrap_or_else(|_| "Файл не найден".to_string());
    println!("Содержимое settings.json:\n{}", file_content);

    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .invoke_handler(tauri::generate_handler![
            get_settings,
            update_settings,
            read_dir,
            rename,
            mkdir,
            mv,
            rm,
            clear_all,
            df,
            upload_file,
            download_and_save,
            download_folder,
            add_sync_folder,
            get_sync_folders,
            toggle_sync_folder,
            remove_sync_folder,
            get_sync_status,
            start_sync,
            stop_sync
        ])
        .run(tauri::generate_context!())
        .expect("error while running Tauri application");
}
