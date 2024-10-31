import { useEffect, useState } from 'react';
import ProjectRow from './ProjectRow';
import useAdminStore from '../../../store';
import HeaderEntry from './HeaderEntry';
import { ProjectSortField } from '../../../enums';
import Button from "../../Button";
import {postRequest} from "../../../api";
import {errorAlert} from "../../../util";

const ProjectsTable = () => {
    const unsortedProjects = useAdminStore((state) => state.projects);
    const fetchProjects = useAdminStore((state) => state.fetchProjects);
    const [projects, setProjects] = useState<Project[]>([]);
    const [guilds, setGuilds] = useState<string[]>([]);
    const [checked, setChecked] = useState<{[key: number]: boolean}>({});
    const [sortState, setSortState] = useState<SortState<ProjectSortField>>({
        field: ProjectSortField.None,
        ascending: true,
    });

    const handleCheckedChange = (e: React.ChangeEvent<HTMLInputElement>, i: number) => {
        setChecked({  // this change of type is to stop React complaining about "a component is changing an uncontrolled input to be controlled"
            ...checked,
            [i]: e.target.checked,
        });
    };

    const updateSort = (field: SortField) => {
        if (sortState.field === field) {
            // If sorted on same field and descending, reset sort state
            if (!sortState.ascending) {
                setSortState({
                    field: ProjectSortField.None,
                    ascending: true,
                });
                setProjects(unsortedProjects);
                return;
            }

            // Otherwise, sort descending
            setSortState({
                field,
                ascending: false,
            });
        } else {
            // If in different sorted state, sort ascending on new field
            setSortState({
                field: field as ProjectSortField,
                ascending: true,
            });
        }
    };

    // On load, fetch projects
    useEffect(() => {
        fetchProjects();
    }, [fetchProjects]);

    // When projects change, update projects and sort
    useEffect(() => {
        // Reset checked state to an object with all indexes false
        setChecked(() => {
            let newChecked: {[key: number]: boolean} = {};
            unsortedProjects.forEach((_, idx) => {
                newChecked[idx] = false;
            });
            return newChecked;
        });

        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        let sortFunc = (a: Project, b: Project) => 0;
        const asc = sortState.ascending ? 1 : -1;
        switch (sortState.field) {
            case ProjectSortField.Name:
                sortFunc = (a, b) => a.name.localeCompare(b.name) * asc;
                break;
            case ProjectSortField.GuildLocation:
                sortFunc = (a, b) => {
                    const guildComparison = a.guild.localeCompare(b.guild) * asc;
                    if (guildComparison !== 0) return guildComparison;  // if guilds are different, just use those to sort
                    return a.location.localeCompare(b.location) * asc;  // otherwise, secondary sort by location
                };
                break;
            case ProjectSortField.Score:
                sortFunc = (a, b) => (a.score - b.score) * asc;
                break;
            case ProjectSortField.Seen:
                sortFunc = (a, b) => (a.seen - b.seen) * asc;
                break;
            case ProjectSortField.Updated:
                sortFunc = (a, b) =>
                    (a.last_activity - b.last_activity) * asc;
                break;
        }
        setProjects(unsortedProjects.sort(sortFunc));
    }, [unsortedProjects, sortState]);

    const bulkHide = async (hide: boolean) => {
        let toHide: string[] = [];
        Object.entries(checked).forEach(([key, value]) => {
            if (value) {
                toHide.push(projects[parseInt(key)].id);
            }
        });

        if (toHide.length === 0) {
            alert('No projects selected!');
            return;
        }

        const res = await postRequest<YesNoResponse>('/project/hide-unhide-many', {ids: toHide, hide: hide});
        if (res.status === 200) {
            alert(`${toHide.length} project(s) ${hide ? 'hidden' : 'unhidden'} successfully!`);
            await fetchProjects();
        } else {
            errorAlert(res);
        }
    }

    useEffect(() => {
        setGuilds(Array.from(new Set(projects.map(p => p.guild))));
    }, [projects]);

    const selectByGuild = () => {
        const selectedGuild = (document.getElementById('guild-select') as HTMLSelectElement).value;
        let toCheck: number[] = [];
        if (selectedGuild === 'all') {
            toCheck = projects.map((_, idx) => idx);
        } else {
            projects.forEach((project, idx) => {
                if (project.guild === selectedGuild) {
                    toCheck.push(idx);
                }
            });
        }

        setChecked(() => {
            let newChecked: {[key: number]: boolean} = {};
            toCheck.forEach((idx) => {
                newChecked[idx] = true;
            });
            return newChecked;
        });
    }

    return (
        <div className="w-full px-8 pb-4">
            <div className="flex flex-row w-full space-x-4 items-center text-center">
                <div>
                    <Button
                        type="primary"
                        square
                        full
                        className="py-2 px-4 rounded-md"
                        onClick={() => {
                            bulkHide(true);
                        }}
                    >
                        Hide Selected
                    </Button>
                </div>
                <div>
                    <Button
                        type="primary"
                        square
                        full
                        className="py-2 px-4 rounded-md"
                        onClick={() => {
                            bulkHide(false);
                        }}
                    >
                        Unhide Selected
                    </Button>
                </div>
                <p className="text-2l text-nowrap">{Object.values(checked).filter(Boolean).length} project(s) currently selected</p>
                <div className="flex flex-nowrap items-center pl-8">
                    <p className="text-2xl mr-2 align-middle">Guild:</p>
                    <select className="rounded-md align-middle" id="guild-select">
                        {guilds.sort().map((guild, idx) => (
                            <option key={idx} value={guild}>{guild}</option>
                        ))}
                        <option key="all-guild" value="all" className="italic">*Select all*</option>
                        <option key="nil-guild" value="nil" className="italic">*Deselect all*</option>
                    </select>
                    <Button type="outline" square className="ml-2 py-2 px-4 rounded-md" onClick={selectByGuild}>
                        Bulk select
                    </Button>
                </div>
            </div>
            <table className="table-fixed w-full text-lg">
                <tbody>
                <tr>
                    <th className="w-12"></th>
                    <HeaderEntry
                        name="Name"
                        updateSort={updateSort}
                        sortField={ProjectSortField.Name}
                        sortState={sortState}
                        align='left'
                    />
                    <HeaderEntry
                        name="Guild"
                        updateSort={updateSort}
                        sortField={ProjectSortField.GuildLocation}
                        sortState={sortState}
                    />
                    <HeaderEntry
                        name="Location"
                        updateSort={updateSort}
                        sortField={ProjectSortField.GuildLocation}
                        sortState={sortState}
                    />
                    <HeaderEntry
                        name="Live Score"
                        updateSort={updateSort}
                        sortField={ProjectSortField.Score}
                        sortState={sortState}
                    />
                    <HeaderEntry
                        name="Seen"
                        updateSort={updateSort}
                        sortField={ProjectSortField.Seen}
                        sortState={sortState}
                    />
                    <HeaderEntry
                        name="Updated"
                        updateSort={updateSort}
                        sortField={ProjectSortField.Updated}
                        sortState={sortState}
                    />
                    <th className="text-right w-24">Actions</th>
                </tr>
                {projects.map((project: Project, idx) => (
                    <ProjectRow
                        key={idx}
                        idx={idx}
                        project={project}
                        checked={checked[idx]}
                        handleCheckedChange={handleCheckedChange}
                    />
                ))}
                </tbody>
            </table>
        </div>
    );
};

export default ProjectsTable;
