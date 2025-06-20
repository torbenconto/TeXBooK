import { useState } from "react"
import TeXBookLogoSM from "../components/Logo"
import { API_URL } from "../config"

const buttonClass =
  "px-12 py-2 rounded-md border border-gray-300 text-gray-800 hover:bg-[#098842] hover:text-white transition-colors duration-200 shadow-sm font-medium hover:cursor-pointer"

function Setup() {
  const [modalOpen, setModalOpen] = useState(false)
  const [modalType, setModalType] = useState<"local" | "cloud" | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  function openModal(type: "local" | "cloud") {
    setModalType(type)
    setError(null)
    setModalOpen(true)
  }

  function closeModal() {
    setModalOpen(false)
    setModalType(null)
    setError(null)
    setLoading(false)
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    const form = e.currentTarget
    const formData = new FormData(form)
    const name = formData.get("name")?.toString().trim() ?? ""
    const path = formData.get("path")?.toString().trim() ?? ""

    if (!name) {
      setError("Name is required")
      setLoading(false)
      return
    }

    if (modalType === "local" && !path) {
      setError("Path is required for local data source")
      setLoading(false)
      return
    }

    try {
      const res = await fetch(API_URL + "/api/data/add", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          type: modalType,
          path: modalType === "local" ? path : undefined,
        }),
      })

      const json = await res.json()
      if (!res.ok) {
        setError(json.error || "Failed to add data source")
      } else {
        alert(json.status || "Data source added successfully!")
        closeModal()
      }
    } catch (err) {
      setError("Network error: " + (err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="w-full h-screen flex flex-col items-center justify-center bg-white text-center px-4">
      <TeXBookLogoSM />
      <h2 className="text-lg text-gray-700 mt-4">
        The free, open-source knowledge management system
      </h2>
      <div className="flex gap-12 mt-8">
        <button className={buttonClass} onClick={() => openModal("local")}>
          Local
        </button>
        <button className={buttonClass} onClick={() => openModal("cloud")}>
          Cloud
        </button>
      </div>

      {modalOpen && (
        <div className="fixed inset-0 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md shadow-lg relative">
            <button
              className="absolute top-2 right-3 text-gray-600 hover:text-gray-900 font-bold text-xl"
              onClick={closeModal}
              aria-label="Close modal"
            >
              &times;
            </button>
            <h3 className="text-xl font-semibold mb-4">
              Create New {modalType === "local" ? "Local" : "Cloud"} Data Source
            </h3>

            {modalType === "cloud" && (
              <p className="mb-4 text-red-600 font-medium">
                Cloud data sources are not supported yet.
              </p>
            )}

            <form onSubmit={handleSubmit}>
              <label className="block text-left mb-2" htmlFor="name">
                Name
              </label>
              <input
                type="text"
                name="name"
                id="name"
                required
                className="w-full border border-gray-300 rounded-md p-2 mb-4"
              />

              {modalType === "local" && (
                <>
                  <label className="block text-left mb-2" htmlFor="path">
                    Path
                  </label>
                  <input
                    type="text"
                    name="path"
                    id="path"
                    required
                    placeholder="/path/to/local/source"
                    className="w-full border border-gray-300 rounded-md p-2 mb-4"
                  />
                </>
              )}

              {error && (
                <p className="text-red-600 mb-4 font-medium" role="alert">
                  {error}
                </p>
              )}

              <button
                type="submit"
                disabled={modalType === "cloud" || loading}
                className={`${
                  modalType === "cloud"
                    ? "bg-gray-400 cursor-not-allowed"
                    : "bg-[#098842] hover:bg-[#076b32] cursor-pointer"
                } text-white px-4 py-2 rounded-md transition-colors`}
              >
                {loading ? "Creating..." : "Create"}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default Setup
