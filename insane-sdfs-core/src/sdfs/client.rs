
use std::io;

use trpl::StreamExt;
use async_trait::async_trait;

use crate::sdfs::types::types::{
    DeleteFileResult, DownloadFileResult, GetFileResult, ListFilesResult, MoveFileResult,
    SdfsConnectionConfig, SdfsResult,
};

#[async_trait]
pub trait SdfsService {
    // File querying methods
    async fn list_files(&self, directory: &str) -> ListFilesResult;

    async fn get_file(&self, path: &str) -> GetFileResult;

    async fn file_exists(&self, path: &str) -> io::Result<bool>;

    async fn download_file(&self, path: &str) -> DownloadFileResult;

    // File manipulation methods
    async fn delete_file(&self, path: &str) -> DeleteFileResult;

    async fn upload_file(
        &self,
        path: &str,
        data: impl StreamExt<Item = Vec<u8>> + Send,
    ) -> io::Result<bool>;

    async fn move_file(
        &self,
        src_path: &str,
        dest_path: &str,
        overwrite_existing: bool,
    ) -> MoveFileResult;

    async fn rename_file(
        &self,
        src_path: &str,
        new_name: &str,
    ) -> MoveFileResult;

    async fn copy_file(
        &self,
        src_path: &str,
        dest_path: &str,
        overwrite_existing: bool,
    ) -> MoveFileResult;

    async fn create_directory(
        &self,
        path: &str,
        recursive: bool,
    ) -> io::Result<bool>;

    async fn ping(&self) -> bool;
}

pub trait SdfsServiceFactory<T: SdfsService> {
    fn connect<'a>(
        &'a self,
        config: &SdfsConnectionConfig,
    ) -> impl Future<Output = SdfsResult<T>> + 'a;
}
