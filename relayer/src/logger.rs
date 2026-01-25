use std::io::{self, Write};
use log::{Level, LevelFilter, Log, Metadata, Record, SetLoggerError};

/// Simple logger that writes to stdout/stderr
struct SimpleLogger;

/// Initialize the simple stdout/stderr logger
/// Default to INFO level, can be overridden with RUST_LOG env var
pub fn init() -> Result<(), SetLoggerError> {
    let level = std::env::var("RUST_LOG")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(LevelFilter::Info);
    
    log::set_logger(&SimpleLogger)?;
    log::set_max_level(level);
    Ok(())
}

impl Log for SimpleLogger {
    fn enabled(&self, _metadata: &Metadata) -> bool {
        // The log crate handles filtering via set_max_level, so we just accept all
        true
    }

    fn log(&self, record: &Record) {
        match record.level() {
            Level::Error | Level::Warn => {
                let _ = writeln!(io::stderr(), "[{}] {}", record.level(), record.args());
            }
            Level::Info | Level::Debug | Level::Trace => {
                let _ = writeln!(io::stdout(), "[{}] {}", record.level(), record.args());
            }
        }
    }

    fn flush(&self) {
        let _ = io::stdout().flush();
        let _ = io::stderr().flush();
    }
}
