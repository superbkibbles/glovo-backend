package server

import (
	"context"

	filepb "github.com/mendmzury/food-delivery/proto/file"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileServer struct {
	filepb.UnimplementedFileServiceServer
}

func NewFileServer() *FileServer {
	return &FileServer{}
}

func (s *FileServer) UploadFile(ctx context.Context, req *filepb.UploadFileRequest) (*filepb.File, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *FileServer) GetFile(ctx context.Context, req *filepb.GetFileRequest) (*filepb.File, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *FileServer) ListFiles(ctx context.Context, req *filepb.ListFilesRequest) (*filepb.ListFilesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *FileServer) DeleteFile(ctx context.Context, req *filepb.DeleteFileRequest) (*commonpb.SuccessResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *FileServer) GetFileURL(ctx context.Context, req *filepb.GetFileURLRequest) (*filepb.GetFileURLResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}

func (s *FileServer) GetFileContent(ctx context.Context, req *filepb.GetFileContentRequest) (*filepb.GetFileContentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "stub")
}
