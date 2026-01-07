import { useState } from "react";
import { axiosInstance } from "@/lib/axios";
import { toast } from "sonner";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";

const VIBE_OPTIONS = [
  { value: "random", label: "Surprise me" },
  { value: "professional", label: "Professional" },
  { value: "creative", label: "Creative" },
  { value: "techy", label: "Techy" },
  { value: "elegant", label: "Elegant" },
  { value: "quirky", label: "Quirky" },
] as const;

const GenerativeIdentityPage = () => {
  const [selectedVibe, setSelectedVibe] = useState("anything");
  const [count, setCount] = useState("5");
  const [usernames, setUsernames] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const handleGenerate = async () => {
    setIsLoading(true);
    try {
      const response = await axiosInstance.post("/api/identity/username", {
        vibe: selectedVibe,
        count: count,
      });

      // Parse CSV response into array, trim whitespace
      const parsedUsernames = response.data.usernames
        .split(",")
        .map((name: string) => name.trim())
        .filter((name: string) => name.length > 0);

      setUsernames(parsedUsernames);
    } catch (error) {
      toast.error("Failed to generate usernames");
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard!");
  };

  const copyAllUsernames = () => {
    navigator.clipboard.writeText(usernames.join("\n"));
    toast.success("All usernames copied!");
  };

  return (
    <div className="space-y-6">

      <Card>
        <CardHeader>
          <CardTitle>Username Generator</CardTitle>
          <CardDescription>
            Select a vibe and how many usernames you want.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Selection controls */}
          <div className="flex flex-wrap gap-4">
            {/* Vibe Select */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Vibe</label>
              <Select value={selectedVibe} onValueChange={setSelectedVibe}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select vibe" />
                </SelectTrigger>
                <SelectContent>
                  {VIBE_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Count Select */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Count</label>
              <Select value={count} onValueChange={setCount}>
                <SelectTrigger className="w-[100px]">
                  <SelectValue placeholder="Count" />
                </SelectTrigger>
                <SelectContent>
                  {Array.from({ length: 10 }, (_, i) => (
                    <SelectItem key={i + 1} value={(i + 1).toString()}>
                      {i + 1}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Generate Button */}
          <Button onClick={handleGenerate} disabled={isLoading}>
            {isLoading ? "Generating..." : "Generate Usernames"}
          </Button>

          {/* Results Section */}
          {usernames.length > 0 && (
            <div className="space-y-3 pt-4 border-t">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Generated Usernames</span>
                <Button variant="outline" size="sm" onClick={copyAllUsernames}>
                  Copy All
                </Button>
              </div>

              <div className="grid gap-2">
                {usernames.map((username, index) => (
                  <div
                    key={index}
                    onClick={() => copyToClipboard(username)}
                    className="p-3 rounded-md border bg-muted/50 hover:bg-muted cursor-pointer transition-colors"
                  >
                    <span className="font-mono text-sm">{username}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default GenerativeIdentityPage;