# For online document, you can use the following code:
import logging
import re
import urllib.parse
from collections.abc import Generator
from typing import Any

from dify_plugin.entities.datasource import (
    DatasourceGetPagesResponse,
    DatasourceMessage,
    GetOnlineDocumentPageContentRequest,
    OnlineDocumentInfo,
    OnlineDocumentPage,
)
from dify_plugin.interfaces.datasource.online_document import OnlineDocumentDatasource

logger = logging.getLogger(__name__)


class {{ .PluginName | SnakeToCamel }}DataSource(OnlineDocumentDatasource):

    def _get_pages(self, datasource_parameters: dict[str, Any]) -> DatasourceGetPagesResponse:
        page = OnlineDocumentPage(
            page_name="",
            page_id="",
            type="page",
            last_edited_time="",
            parent_id="",
            page_icon=None,
        )
        # Get workspace info from credentials
        workspace_name = self.runtime.credentials.get("workspace_name", "{{ .PluginName }}")
        workspace_id = self.runtime.credentials.get("workspace_id", "unknown")
        workspace_icon = self.runtime.credentials.get("workspace_icon", "")

        online_document_info = OnlineDocumentInfo(
            workspace_name=workspace_name,
            workspace_icon=workspace_icon,
            workspace_id=workspace_id,
            pages=[page],
            total=1,
        )
        return DatasourceGetPagesResponse(result=[online_document_info])

    def _get_content(self, page: GetOnlineDocumentPageContentRequest) -> Generator[DatasourceMessage, None, None]:
       
        yield self.create_variable_message("content", "")
        yield self.create_variable_message("page_id", "")
        yield self.create_variable_message("workspace_id", "")


# For website crawl, you can use the following code:
# from typing import Any, Generator

# from dify_plugin.entities.datasource import WebSiteInfo, WebSiteInfoDetail
# from dify_plugin.entities.tool import ToolInvokeMessage
# from dify_plugin.interfaces.datasource.website import WebsiteCrawlDatasource


# class {{ .PluginName | SnakeToCamel }}DataSource(WebsiteCrawlDatasource):

#     def _get_website_crawl(
#         self, datasource_parameters: dict[str, Any]
#     ) -> Generator[ToolInvokeMessage, None, None]:

#         crawl_res = WebSiteInfo(web_info_list=[], status="", total=0, completed=0)
#         crawl_res.status = "processing"
#         yield self.create_crawl_message(crawl_res)

#         crawl_res.status = "completed"
#         crawl_res.web_info_list = [
#             WebSiteInfoDetail(
#                 title="",
#                 source_url="",
#                 description="",
#                 content="",
#             )
#         ]
#         crawl_res.total = 1
#         crawl_res.completed = 1

#         yield self.create_crawl_message(crawl_res)


# For online drive, you can use the following code:
# from collections.abc import Generator

# from dify_plugin.entities.datasource import (
#     DatasourceMessage,
#     OnlineDriveBrowseFilesRequest,
#     OnlineDriveBrowseFilesResponse,
#     OnlineDriveDownloadFileRequest,
#     OnlineDriveFile,
#     OnlineDriveFileBucket,
# )
# from dify_plugin.interfaces.datasource.online_drive import OnlineDriveDatasource


# class {{ .PluginName | SnakeToCamel }}DataSource(OnlineDriveDatasource):

#     def _browse_files(
#         self, request: OnlineDriveBrowseFilesRequest
#     ) -> OnlineDriveBrowseFilesResponse:

#         credentials = self.runtime.credentials
#         bucket_name = request.bucket
#         prefix = request.prefix or ""  # Allow empty prefix for root folder; When you browse the folder, the prefix is the folder id
#         max_keys = request.max_keys or 10
#         next_page_parameters = request.next_page_parameters or {}

#         files = []
#         files.append(OnlineDriveFile(
#             id="", 
#             name="", 
#             size=0, 
#             type="folder" # or "file"
#         ))

#         return OnlineDriveBrowseFilesResponse(result=[
#             OnlineDriveFileBucket(
#                 bucket="", 
#                 files=files, 
#                 is_truncated=False, 
#                 next_page_parameters={}
#             )
#         ])

#     # if file.type is "file", the plugin will download the file content
#     def _download_file(self, request: OnlineDriveDownloadFileRequest) -> Generator[DatasourceMessage, None, None]:
#         credentials = self.runtime.credentials
#         file_id = request.id

#         file_content = bytes()
#         file_name = ""

#         mime_type = self._get_mime_type_from_filename(file_name)
        
#         yield self.create_blob_message(file_content, meta={
#             "file_name": file_name,
#             "mime_type": mime_type
#         })

#     def _get_mime_type_from_filename(self, filename: str) -> str:
#         """Determine MIME type from file extension."""
#         import mimetypes
#         mime_type, _ = mimetypes.guess_type(filename)
#         return mime_type or "application/octet-stream"