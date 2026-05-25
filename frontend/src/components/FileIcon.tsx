import FolderIcon from "@mui/icons-material/Folder";
import InsertDriveFileIcon from "@mui/icons-material/InsertDriveFile";
import ImageIcon from "@mui/icons-material/Image";
import MovieIcon from "@mui/icons-material/Movie";
import AudioFileIcon from "@mui/icons-material/AudioFile";
import DescriptionIcon from "@mui/icons-material/Description";
import ArticleIcon from "@mui/icons-material/Article";
import ArchiveIcon from "@mui/icons-material/Archive";
import CodeIcon from "@mui/icons-material/Code";
import type { SvgIconProps } from "@mui/material";

interface Props extends SvgIconProps {
  isDir: boolean;
  name: string;
}

const imageExt = /\.(png|jpe?g|gif|webp|bmp|svg|ico|tiff?)$/i;
const videoExt = /\.(mp4|mkv|mov|avi|webm|flv|wmv)$/i;
const audioExt = /\.(mp3|wav|ogg|flac|m4a|aac)$/i;
const docExt = /\.(pdf|docx?|odt|rtf)$/i;
const sheetExt = /\.(xlsx?|csv|ods)$/i;
const archiveExt = /\.(zip|tar|gz|bz2|xz|7z|rar)$/i;
const codeExt = /\.(go|rs|ts|tsx|js|jsx|py|rb|java|c|cpp|h|hpp|sh|md|json|yaml|yml|toml)$/i;

export default function FileIcon({ isDir, name, ...rest }: Props) {
  if (isDir) return <FolderIcon color="primary" {...rest} />;
  if (imageExt.test(name)) return <ImageIcon color="success" {...rest} />;
  if (videoExt.test(name)) return <MovieIcon color="info" {...rest} />;
  if (audioExt.test(name)) return <AudioFileIcon color="warning" {...rest} />;
  if (docExt.test(name)) return <DescriptionIcon color="error" {...rest} />;
  if (sheetExt.test(name)) return <ArticleIcon color="success" {...rest} />;
  if (archiveExt.test(name)) return <ArchiveIcon color="warning" {...rest} />;
  if (codeExt.test(name)) return <CodeIcon color="secondary" {...rest} />;
  return <InsertDriveFileIcon {...rest} />;
}
