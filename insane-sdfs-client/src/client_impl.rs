use std::io;

use insane_sdfs_core::sdfs::client::SdfsService;
use insane_sdfs_core::sdfs::types::types::ListFilesResult;

pub struct TcpSdfsClient {}

#[async_trait::async_trait]
impl SdfsService for TcpSdfsClient {
    async fn list_files(&self, _directory: &str) -> ListFilesResult {
        unimplemented!()
    }

    async fn get_file(&self, _path: &str) -> insane_sdfs_core::sdfs::types::types::GetFileResult {
        unimplemented!()
    }

    async fn file_exists(&self, _path: &str) -> io::Result<bool> {
        unimplemented!()
    }

    async fn download_file(
        &self,
        _path: &str,
    ) -> insane_sdfs_core::sdfs::types::types::DownloadFileResult {
        unimplemented!()
    }

    async fn delete_file(
        &self,
        _path: &str,
    ) -> insane_sdfs_core::sdfs::types::types::DeleteFileResult {
        unimplemented!()
    }

    async fn upload_file(
        &self,
        _path: &str,
        _data: impl trpl::StreamExt<Item = Vec<u8>> + Send,
    ) -> io::Result<bool> {
        unimplemented!()
    }

    async fn move_file(
        &self,
        _src_path: &str,
        _dest_path: &str,
        _overwrite_existing: bool,
    ) -> insane_sdfs_core::sdfs::types::types::MoveFileResult {
        unimplemented!()
    }

    async fn rename_file(
        &self,
        _src_path: &str,
        _new_name: &str,
    ) -> insane_sdfs_core::sdfs::types::types::MoveFileResult {
        unimplemented!()
    }

    async fn copy_file(
        &self,
        _src_path: &str,
        _dest_path: &str,
        _overwrite_existing: bool,
    ) -> insane_sdfs_core::sdfs::types::types::MoveFileResult {
        unimplemented!()
    }

    async fn create_directory(&self, _path: &str, _recursive: bool) -> io::Result<bool> {
        unimplemented!()
    }

    async fn ping(&self) -> bool {
        unimplemented!()
    }
}
