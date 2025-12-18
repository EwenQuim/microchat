import { useNavigate } from "@tanstack/react-router";
import { Command } from "cmdk";
import { Search } from "lucide-react";
import { useEffect, useState } from "react";
import { useRooms } from "@/hooks/useRooms";
import { useSearchRooms } from "@/hooks/useSearchRooms";

interface SearchCommandProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function SearchCommand({ open, onOpenChange }: SearchCommandProps) {
	const [search, setSearch] = useState("");
	const navigate = useNavigate();
	const { data: searchResults } = useSearchRooms(search);
	const { data: allRooms } = useRooms();

	const rooms = search.trim() ? searchResults : allRooms;

	useEffect(() => {
		const down = (e: KeyboardEvent) => {
			if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
				e.preventDefault();
				onOpenChange(!open);
			}
			if (e.key === "Escape" && open) {
				e.preventDefault();
				onOpenChange(false);
			}
		};

		document.addEventListener("keydown", down);
		return () => document.removeEventListener("keydown", down);
	}, [open, onOpenChange]);

	const handleSelect = (roomName: string) => {
		navigate({ to: "/chat/$roomName", params: { roomName } });
		onOpenChange(false);
		setSearch("");
	};

	if (!open) return null;

	return (
		<button
			type="button"
			className="fixed inset-0 z-50 bg-black/50 cursor-default"
			onClick={() => onOpenChange(false)}
		>
			<div className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-lg">
				<Command
					className="bg-background rounded-lg border shadow-lg"
					onClick={(e) => e.stopPropagation()}
				>
					<div className="flex items-center border-b px-3">
						<Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
						<Command.Input
							placeholder="Search rooms..."
							value={search}
							onValueChange={setSearch}
							className="flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
						/>
					</div>
					<Command.List className="max-h-[300px] overflow-y-auto p-2">
						<Command.Empty className="py-6 text-center text-sm text-muted-foreground">
							No rooms found.
						</Command.Empty>
						{rooms
							?.filter((room) => !room.hidden)
							.map((room) => (
								<Command.Item
									key={room.name}
									value={room.name}
									onSelect={() => handleSelect(room.name || "")}
									className="relative flex cursor-pointer select-none items-center rounded-sm px-2 py-3 text-sm outline-none hover:bg-accent data-[selected=true]:bg-accent"
								>
									<div className="flex flex-col flex-1">
										<div className="flex items-center justify-between">
											<span className="font-medium">{room.name}</span>
											{room.message_count && room.message_count > 0 && (
												<span className="text-xs text-muted-foreground ml-2">
													{room.message_count} messages
												</span>
											)}
										</div>
										{room.last_message_content && (
											<div className="text-xs text-muted-foreground mt-1 truncate">
												{room.last_message_user && (
													<span className="font-medium">
														{room.last_message_user}:{" "}
													</span>
												)}
												{room.last_message_content}
											</div>
										)}
									</div>
								</Command.Item>
							))}
					</Command.List>
				</Command>
			</div>
		</button>
	);
}
