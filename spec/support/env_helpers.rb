module EnvHelpers
  def stub_env(variable_name, value)
    allow(ENV).to receive(:[]).and_call_original
    allow(ENV).to receive(:[]).with(variable_name).and_return(value)
    allow(ENV).to receive(:fetch).and_call_original
    allow(ENV).to receive(:fetch).with(variable_name).and_return(value)
  end

  RSpec.configure { |config| config.include self }
end
