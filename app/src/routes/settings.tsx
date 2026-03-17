import {
	createFileRoute,
	Link,
	Outlet,
	useLocation,
	useNavigate,
} from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/settings")({
	component: SettingsLayout,
});

const TABS = [
	{ value: "user", label: "User", to: "/settings/user" },
	{ value: "identities", label: "Identities", to: "/settings/identities" },
	{ value: "import", label: "Import", to: "/settings/import" },
	{ value: "export", label: "Export", to: "/settings/export" },
	{ value: "contacts", label: "Contacts", to: "/settings/contacts" },
	{ value: "servers", label: "Servers", to: "/settings/servers" },
] as const;

function SettingsLayout() {
	const navigate = useNavigate();
	const { pathname } = useLocation();

	return (
		<div className="min-h-screen bg-background text-foreground p-4 md:p-8">
			<div className="max-w-4xl mx-auto">
				<div className="flex items-center gap-4 mb-8">
					<Button
						onClick={() => navigate({ to: "/", search: {} })}
						variant="ghost"
						size="icon"
						className="text-muted-foreground hover:text-foreground h-12 w-12 md:h-10 md:w-10"
					>
						<ArrowLeft className="h-6 w-6 md:h-5 md:w-5" />
					</Button>
					<h1 className="text-xl md:text-3xl font-bold">Settings</h1>
				</div>

				<div className="flex gap-2 mb-6 border-b border-border overflow-x-auto -mx-4 md:-mx-8 px-4 md:px-8">
					{TABS.map((tab) => (
						<Link
							key={tab.value}
							to={tab.to}
							className={`px-4 md:px-6 py-3 font-medium transition-colors whitespace-nowrap ${
								pathname === tab.to
									? "border-b-2 border-cyan-500 text-cyan-500"
									: "text-muted-foreground hover:text-foreground"
							}`}
						>
							{tab.label}
						</Link>
					))}
				</div>

				<div className="bg-card text-card-foreground rounded-lg p-4 md:p-6 border border-border">
					<Outlet />
				</div>
			</div>
		</div>
	);
}
