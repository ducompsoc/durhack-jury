import { twMerge } from 'tailwind-merge';
import DragHamburger from './dnd/DragHamburger';
import {truncate} from "../../util";

interface ProjectEntryProps {
    project: SortableJudgedProject;
    ranking: number;
}

const ProjectEntry = ({ project, ranking }: ProjectEntryProps) => {
    if (!project) {
        return null;
    }

    let rankColor = 'text-lightest';
    switch (ranking) {
        case 1:
            rankColor = 'text-gold';
            break;
        case 2:
            rankColor = 'text-light';
            break;
        case 3:
            rankColor = 'text-bronze';
            break;
        default:
            break;
    }

    return (
        <div className="flex items-center cursor-default">
            {ranking !== -1 && (
                <p className={twMerge('font-bold text-xl text-center w-6 shrink-0', rankColor)}>{ranking}</p>
            )}
            <div className="m-1 pl-2 py-1 bg-background border-solid border-2 border-lightest rounded-md grow min-w-0">
                <div className="flex flex-row">
                    <div className="min-w-0">
                        <h3 className="text-xl grow">
                            <a href={`/judge/project/${project.project_id}`}>
                                <b>{truncate(project.name, 15)}</b>&nbsp;({truncate(project.guild, 5)}|{truncate(project.location, 10)})
                            </a>
                        </h3>
                        <p className="text-light text-xs line-clamp-1">{project.notes}</p>
                        <div className="text-light flex flex-wrap">
                            {Object.entries(project.categories).map(([name, score], i) => (
                                <div key={i}>
                                    <span className="text-lighter text-xs mr-1">
                                        {truncate(name, 15)}
                                    </span>
                                    <span className="mr-2">{score}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                    <div className="grow text-right flex items-center justify-end">
                        <DragHamburger />
                        {/* <button onClick={openProject} className="text-3xl w-10 h-10 font-bold p-2 text-light duration-200 hover:text-primary leading-[0.5] rounded-full">
                        +
                    </button> */}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ProjectEntry;
