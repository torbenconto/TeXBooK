import { useEffect, useState } from "react";
import Setup from "./Setup";
import TeXBookLogoSM from "../components/LogoSM";
import { API_URL } from "../config";
import { Document, Page, pdfjs } from "react-pdf";

pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.mjs`;

type FileEntry = {
  path: string;
  thumbnail?: string | null;
};

type DataSource = {
  name: string;
  metadata: any;
};

export default function Home() {
  const [data, setData] = useState<DataSource[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSource, setSelectedSource] = useState<string | null>(null);
  const [files, setFiles] = useState<FileEntry[] | null>(null);
  const [filesLoading, setFilesLoading] = useState(false);
  const [filesError, setFilesError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        const res = await fetch(`${API_URL}/api/data/list`);
        if (!res.ok) throw new Error(`Error ${res.status}: ${res.statusText}`);
        const json = await res.json();
        setData(json);

        if (json.length > 0) {
          setSelectedSource(json[0].name || json[0].id || "");
        }
      } catch (err: any) {
        setError(err.message || "Unknown error");
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  useEffect(() => {
    if (!selectedSource) {
      setFiles(null);
      return;
    }

    setFilesLoading(true);
    setFilesError(null);

    async function fetchFiles() {
      try {
        const res = await fetch(
          `${API_URL}/api/fs/list?source=${encodeURIComponent(selectedSource!)}`
        );
        if (!res.ok) throw new Error(`Error ${res.status}: ${res.statusText}`);
        const json = await res.json();
        setFiles(json.files || []);
      } catch (err: any) {
        setFilesError(err.message || "Unknown error fetching files");
      } finally {
        setFilesLoading(false);
      }
    }

    fetchFiles();
  }, [selectedSource]);

  if (loading) return <p>Loading data sources...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;

  if (data && data.length === 0) {
    return <Setup />;
  }

  return (
    <div className="min-h-screen bg-white flex flex-col">
      <header className="w-full flex items-center justify-center relative px-4 py-4 border-b border-gray-200">
        <TeXBookLogoSM />
        <div className="absolute right-4 flex items-center gap-2">
          <label
            htmlFor="data-source-select"
            className="text-sm font-medium text-gray-700"
          >
            Source:
          </label>
          <select
            id="data-source-select"
            value={selectedSource ?? ""}
            onChange={(e) => setSelectedSource(e.target.value)}
            className="border border-gray-300 rounded-md p-1"
          >
            {data?.map((source) => (
              <option key={source.name} value={source.name}>
                {source.name}
              </option>
            ))}
          </select>
        </div>
      </header>

      <main className="flex-grow p-8">
        {filesLoading && <p>Loading files...</p>}
        {filesError && <p className="text-red-600">Error: {filesError}</p>}

        {!filesLoading && !filesError && files && files.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {files.slice(0, 10).map(({ path, thumbnail }) => {
              if (!thumbnail) return null;
              const fileUrl = `${API_URL}/api/fs/file?source=${encodeURIComponent(
                selectedSource!
              )}&path=${encodeURIComponent(thumbnail)}`;

              return (
                <div key={path} className="flex flex-col items-center">
                  <div className="w-48 aspect-[5/9] bg-gray-100 flex items-center justify-center overflow-hidden rounded-md">
                    <Document
                      file={fileUrl}
                      loading={
                        <span className="text-gray-400 text-sm">
                          Loading PDF...
                        </span>
                      }
                      error={
                        <span className="text-red-600 text-sm">
                          Error loading PDF
                        </span>
                      }
                    >
                      <Page
                        pageNumber={1}
                        width={192}
                        scale={2}
                        renderTextLayer={false}
                        renderAnnotationLayer={false}
                        renderMode="canvas"
                      />
                    </Document>
                  </div>
                  <p className="text-sm break-words text-center mt-2 px-1">{path}</p>
                </div>
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
}
