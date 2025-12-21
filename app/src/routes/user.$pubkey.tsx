import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import { hexToNpub } from "@/lib/crypto";

export const Route = createFileRoute("/user/$pubkey")({
	component: UserProfilePage,
});

interface UserWithPostCount {
	public_key: string;
	verified: boolean;
	created_at: string;
	updated_at: string;
	post_count: number;
}

async function fetchUserDetails(publicKey: string): Promise<UserWithPostCount> {
	const res = await fetch(`/api/users/${publicKey}/details`);
	if (!res.ok) {
		throw new Error("Failed to fetch user details");
	}
	const body = await res.text();
	return JSON.parse(body);
}

function UserProfilePage() {
	const { pubkey } = Route.useParams();

	const {
		data: user,
		isLoading,
		error,
	} = useQuery({
		queryKey: ["user-details", pubkey],
		queryFn: () => fetchUserDetails(pubkey),
	});

	if (isLoading) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center">Loading user profile...</div>
			</div>
		);
	}

	if (error || !user) {
		return (
			<div className="container mx-auto p-8">
				<div className="text-center text-red-500">
					Error loading user profile: {error?.message || "User not found"}
				</div>
			</div>
		);
	}

	const createdDate = formatDistanceToNow(new Date(user.created_at), {
		addSuffix: true,
	});

	const npub = hexToNpub(user.public_key);

	return (
		<div className="container mx-auto p-8">
			<div className="max-w-2xl mx-auto">
				<h1 className="text-3xl font-bold mb-6">User Profile</h1>

				<div className="bg-muted rounded-lg p-6 space-y-4">
					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Public Key (npub)
						</div>
						<p className="font-mono text-sm break-all mt-1">{npub}</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Verified by {window.location.hostname} admins
						</div>
						<p className="mt-1">
							{user.verified ? (
								<span className="text-green-500">✓ Verified</span>
							) : (
								<span className="text-muted-foreground">Not verified</span>
							)}
						</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Posts on this server
						</div>
						<p className="text-2xl font-bold mt-1">{user.post_count}</p>
					</div>

					<div>
						<div className="text-sm font-semibold text-muted-foreground">
							Member since
						</div>
						<p className="mt-1">{createdDate}</p>
					</div>
				</div>
			</div>
		</div>
	);
}
