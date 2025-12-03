
use chrono::{DateTime, Utc};
use trpl::Stream;

use crate::sdfs::client::SdfsService;

// Type aliases for clarity
pub type Token = String;
pub type Buffer = Vec<u8>;

// Configuration for connecting to an SDFS server
pub struct SdfsConnectionConfig {
    pub address: String,
    pub port: u16,
    pub token: Token,
}

// Representation of a file in SDFS
pub enum SdfsFile {
    RegularFile {
        name: String,
        size: u64,
        created_at: DateTime<Utc>,
        last_modified: DateTime<Utc>,
    },
    Directory {
        name: String,
        created_at: DateTime<Utc>,
        last_modified: DateTime<Utc>,
    },
}

impl SdfsFile {
    pub fn get_parent_directory_path(&self) -> Option<&str> {
        match self {
            SdfsFile::RegularFile { name, .. } | SdfsFile::Directory { name, .. } => {
                std::path::Path::new(name).parent().and_then(|p| p.to_str())
            }
        }
    }
}

// Result type for SDFS operations
pub enum SdfsResult<T: SdfsService> {
    Success(T),
    HostUnreachable,
    ProtocolError,
    AuthenticationFailed,
}

pub enum ListFilesResult {
    Success(Vec<SdfsFile>),
    DirectoryNotFound,
    FileNotDirectory,
}

pub enum GetFileResult {
    Success(SdfsFile),
    FileNotFound,
}

pub enum DownloadFileResult {
    Success(Box<dyn Stream<Item = Buffer> + Unpin>),
    FileNotFound,
    FileIsNotRegular,
}

pub enum DeleteFileResult {
    Success,
    FileNotFound,
    DirectoryNotEmpty,
}

pub enum MoveFileResult {
    Success,
    FileNotFound,
    FileAlreadyExists,
}
