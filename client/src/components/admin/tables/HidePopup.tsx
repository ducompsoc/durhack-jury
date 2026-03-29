import { Project } from "../../../types";

interface HideProjectPopupProps {
    /* Project to edit */
    project: Project;

    /* Function to modify the popup state variable */
    close: React.Dispatch<React.SetStateAction<boolean>>;
}

const HideProjectPopup = ({ project, close }: HideProjectPopupProps) => {
    const options = ['Guild sent to lunch', 'Team not found', 'Other'];
    return (
        <>
            <div className="bg-background fixed z-20 left-1/2 top-1/2 translate-x-[-50%] translate-y-[-50%] py-6 px-10 w-1/3">
                <h1 className="text-5xl font-bold mb-2 text-center">Enter a reason for hiding {project.name}</h1>
                <div className="flex flex-col">
                    <ul className="list-disc list-inside">
                        {options.map((option) => <li key={option}>{option}</li>)}
                    </ul>
                </div>
                <div className="flex flex-row justify-around">
                    <button
                        className=" border-lightest border-2 rounded-full px-6 py-1 mt-4 w-2/5 font-bold text-2xl text-lighter hover:bg-lighter/30 duration-200"
                        onClick={() => close(false)}
                    >
                        Cancel
                    </button>
                    <button
                        className="bg-primary rounded-full px-6 py-1 mt-4 w-2/5 font-bold text-2xl text-background hover:bg-primary/80 duration-200"
                        onClick={() => close(false)}
                    >
                        Submit
                    </button>
                </div>
            </div>
        </>
    );
};

export default HideProjectPopup;
