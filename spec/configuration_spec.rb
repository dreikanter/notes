require_relative "../src/configuration"

RSpec.describe Configuration do
  subject(:configuration) { described_class.new(file_name) }

  let(:file_name) { file_fixture("configuration/sample.yml") }

  it { expect(configuration.notes_path).to eq("/home/user/notes") }
  it { expect(configuration.build_path).to eq(File.join(file_fixture_path, "configuration/dist")) }
  it { expect(configuration.templates_path).to eq(File.join(file_fixture_path, "configuration/templates")) }
  it { expect(configuration.site_root_url).to eq("https://notes.musayev.com") }
  it { expect(configuration.site_name).to eq("Alex Musayev Notes") }
  it { expect(configuration.author_name).to eq("Alex Musayev") }
  it { expect(configuration.site_root_path).to eq("/") }
end
