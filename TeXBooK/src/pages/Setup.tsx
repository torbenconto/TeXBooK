import TeXBookLogo from "../components/Logo"

const buttonClass = "px-12 py-2 rounded-md border border-gray-300 text-gray-800 hover:bg-[#098842] hover:text-white transition-colors duration-200 shadow-sm font-medium"

function Setup() {
    return (
        <div className="w-full h-screen flex flex-col items-center justify-center bg-white text-center px-4">
        <TeXBookLogo />
        <h2 className="text-lg text-gray-700 mt-4">
            The free, open-source knowledge management system
        </h2>
        <div className="flex gap-12 mt-8">
            <button className={buttonClass}>Local</button>
            <button className={buttonClass}>Cloud</button>
        </div>
        </div>
    )
}

export default Setup