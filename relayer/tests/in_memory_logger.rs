use log::{Level, Log, Metadata, Record, SetLoggerError};
use std::sync::Mutex;

static ERRORS: Mutex<Vec<String>> = Mutex::new(Vec::new());
static WARNS: Mutex<Vec<String>> = Mutex::new(Vec::new());
static INFOS: Mutex<Vec<String>> = Mutex::new(Vec::new());
static DEBUGS: Mutex<Vec<String>> = Mutex::new(Vec::new());
static TRACES: Mutex<Vec<String>> = Mutex::new(Vec::new());

static LOGGER: InMemoryLogger = InMemoryLogger;

/// Tests should run sequentially (--test-threads=1) to safely use clear()
/// We can get rid of mutexes if we create an interface that is received by lib::run()
pub struct InMemoryLogger;

impl Log for InMemoryLogger {
    fn enabled(&self, _metadata: &Metadata) -> bool {
        true
    }

    fn log(&self, record: &Record) {
        let message = format!("{}", record.args());
        match record.level() {
            Level::Error => {
                ERRORS.lock().unwrap().push(message);
            }
            Level::Warn => {
                WARNS.lock().unwrap().push(message);
            }
            Level::Info => {
                INFOS.lock().unwrap().push(message);
            }
            Level::Debug => {
                DEBUGS.lock().unwrap().push(message);
            }
            Level::Trace => {
                TRACES.lock().unwrap().push(message);
            }
        }
    }

    fn flush(&self) {
        // No-op for in-memory logger
    }
}

impl InMemoryLogger {
    pub fn init(level: Level) -> Result<Self, SetLoggerError> {
        log::set_logger(&LOGGER)?;
        log::set_max_level(level.to_level_filter());
        Ok(InMemoryLogger)
    }

    pub fn errors(&self) -> Vec<String> {
        ERRORS.lock().unwrap().clone()
    }

    pub fn warns(&self) -> Vec<String> {
        WARNS.lock().unwrap().clone()
    }

    pub fn infos(&self) -> Vec<String> {
        INFOS.lock().unwrap().clone()
    }

    pub fn debugs(&self) -> Vec<String> {
        DEBUGS.lock().unwrap().clone()
    }

    pub fn traces(&self) -> Vec<String> {
        TRACES.lock().unwrap().clone()
    }

    // Safe to clear when tests run sequentially (--test-threads=1)
    pub fn clear() {
        ERRORS.lock().unwrap().clear();
        WARNS.lock().unwrap().clear();
        INFOS.lock().unwrap().clear();
        DEBUGS.lock().unwrap().clear();
        TRACES.lock().unwrap().clear();
    }
}
