import { useSearchParams, useParams } from "react-router-dom";
import { API_URL } from "../config";

function Viewer() {
  const { source } = useParams();
  const [searchParams] = useSearchParams();
  const path = searchParams.get("path");

  const fileUrl = `${API_URL}/api/v1/datasources/${source}/fs/file?path=${encodeURIComponent(path ?? "")}`;

  return (
    <main className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
      {path ? (
        <iframe
          src={fileUrl}
          title="PDF Viewer"
          className="h-screen w-screen"
        />
      ) : (
        <p className="text-center text-gray-500">No file path provided.</p>
      )}
    </main>
  );
}

export default Viewer;
